package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"
	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

var (
	_ resource.Resource                     = &driftDetectionResource{}
	_ resource.ResourceWithConfigure        = &driftDetectionResource{}
	_ resource.ResourceWithConfigValidators = &driftDetectionResource{}
	_ resource.ResourceWithImportState      = &driftDetectionResource{}
	_ resource.ResourceWithModifyPlan       = &driftDetectionResource{}
)

func newDriftDetectionResource() resource.Resource {
	return &driftDetectionResource{}
}

type driftDetectionResource struct {
	framework.ResourceWithScalrClient
}

type driftDetectionResourceModel struct {
	Id               types.String `tfsdk:"id"`
	EnvironmentID    types.String `tfsdk:"environment_id"`
	CheckPeriod      types.String `tfsdk:"check_period"`
	WorkspaceFilters types.Set    `tfsdk:"workspace_filters"`
	RunMode          types.String `tfsdk:"run_mode"`
}

type workspaceFiltersModel struct {
	NamePatterns     types.Set `tfsdk:"name_patterns"`
	EnvironmentTypes types.Set `tfsdk:"environment_types"`
	Tags             types.Set `tfsdk:"tags"`
}

func (r *driftDetectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drift_detection"
}

func (r *driftDetectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of Drift Detection Scheduler in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"check_period": schema.StringAttribute{
				MarkdownDescription: "Check period for drift detection: daily or weekly.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.DriftDetectionSchedulePeriodDaily),
						string(scalr.DriftDetectionSchedulePeriodWeekly),
					),
				},
			},
			"run_mode": schema.StringAttribute{
				MarkdownDescription: "Run mode for drift detection: `refresh-only` (default) or `plan`. ",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(scalr.DriftDetectionScheduleRunModeRefreshOnly)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.DriftDetectionScheduleRunModeRefreshOnly),
						string(scalr.DriftDetectionScheduleRunModePlan),
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"workspace_filters": schema.SetNestedBlock{
				MarkdownDescription: "Filters for workspaces to be included in drift detection. Only one type of filter can be specified: `name_patterns`, `environment_types` or `tags`.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name_patterns": schema.SetAttribute{
							MarkdownDescription: "Workspace name patterns to include in drift detection. Supports `*` wildcard (e.g., `prod-*`).",
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("environment_types"),
									path.MatchRelative().AtParent().AtName("tags"),
								),
							},
						},
						"environment_types": schema.SetAttribute{
							MarkdownDescription: "Workspace environment types to include in drift detection. Allowed values: `production`, `staging`, `testing`, `development`, `unmapped`.",
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("name_patterns"),
									path.MatchRelative().AtParent().AtName("tags"),
								),
							},
						},
						"tags": schema.SetAttribute{
							MarkdownDescription: "Workspace tag to include in drift detection. A workspace matches if it has at least one of the specified tags.",
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("name_patterns"),
									path.MatchRelative().AtParent().AtName("environment_types"),
								),
							},
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func (r *driftDetectionResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func toScalrWorkspaceFilters(ctx context.Context, tfSet types.Set) (*scalr.DriftDetectionWorkspaceFilter, diag.Diagnostics) {
	var diags diag.Diagnostics
	var filters []workspaceFiltersModel

	scalrWorkspaceFilter := &scalr.DriftDetectionWorkspaceFilter{}

	if tfSet.IsNull() || tfSet.IsUnknown() {
		return scalrWorkspaceFilter, diags
	}

	diags.Append(tfSet.ElementsAs(ctx, &filters, false)...)
	if diags.HasError() {
		return nil, diags
	}
	if len(filters) == 0 {
		return scalrWorkspaceFilter, diags
	}
	filter := filters[0]
	if !filter.NamePatterns.IsNull() && !filter.NamePatterns.IsUnknown() {
		var namePatterns []string
		diags.Append(filter.NamePatterns.ElementsAs(ctx, &namePatterns, false)...)
		if diags.HasError() {
			return nil, diags
		}
		scalrWorkspaceFilter.NamePatterns = &namePatterns
	}
	if !filter.EnvironmentTypes.IsNull() && !filter.EnvironmentTypes.IsUnknown() {
		var environmentTypes []scalr.WorkspaceEnvironmentType
		diags.Append(filter.EnvironmentTypes.ElementsAs(ctx, &environmentTypes, false)...)
		if diags.HasError() {
			return nil, diags
		}
		scalrWorkspaceFilter.EnvironmentTypes = &environmentTypes
	}
	if !filter.Tags.IsNull() && !filter.Tags.IsUnknown() {
		var tags []string
		diags.Append(filter.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return nil, diags
		}
		scalrWorkspaceFilter.Tags = &tags
	}
	return scalrWorkspaceFilter, diags
}

func toTerraformWorkspaceFilters(ctx context.Context, scalrWorkspaceFilter scalr.DriftDetectionWorkspaceFilter) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	filters := []workspaceFiltersModel{
		{
			NamePatterns:     types.SetNull(types.StringType),
			EnvironmentTypes: types.SetNull(types.StringType),
			Tags:             types.SetNull(types.StringType),
		},
	}

	if !scalrWorkspaceFilter.IsEmpty() {
		filter := &filters[0]
		if scalrWorkspaceFilter.NamePatterns != nil {
			namePatterns, d := types.SetValueFrom(ctx, types.StringType, *scalrWorkspaceFilter.NamePatterns)
			diags.Append(d...)
			filter.NamePatterns = namePatterns
		} else if scalrWorkspaceFilter.EnvironmentTypes != nil {
			envTypes, d := types.SetValueFrom(ctx, types.StringType, *scalrWorkspaceFilter.EnvironmentTypes)
			diags.Append(d...)
			filter.EnvironmentTypes = envTypes
		} else if scalrWorkspaceFilter.Tags != nil {
			tags, d := types.SetValueFrom(ctx, types.StringType, *scalrWorkspaceFilter.Tags)
			diags.Append(d...)
			filter.Tags = tags
		}
	}

	tfSet, d := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name_patterns":     types.SetType{ElemType: types.StringType},
			"environment_types": types.SetType{ElemType: types.StringType},
			"tags":              types.SetType{ElemType: types.StringType},
		},
	}, filters)
	diags.Append(d...)

	return tfSet, diags
}

func (r *driftDetectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan driftDetectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceFilter, diags := toScalrWorkspaceFilters(ctx, plan.WorkspaceFilters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.DriftDetectionCreateOptions{
		Environment:      &scalr.Environment{ID: plan.EnvironmentID.ValueString()},
		Schedule:         scalr.DriftDetectionSchedulePeriod(plan.CheckPeriod.ValueString()),
		WorkspaceFilters: *workspaceFilter,
	}

	if !plan.RunMode.IsUnknown() && !plan.RunMode.IsNull() {
		opts.RunMode = ptr(scalr.DriftDetectionScheduleRunMode(plan.RunMode.ValueString()))
	}

	driftDetection, err := r.Client.DriftDetections.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating scalr_drift_detection", err.Error())
		return
	}

	plan.Id = types.StringValue(driftDetection.ID)
	plan.EnvironmentID = types.StringValue(driftDetection.Environment.ID)
	plan.CheckPeriod = types.StringValue(string(driftDetection.Schedule))
	plan.RunMode = types.StringValue(string(driftDetection.RunMode))

	tfWorkspaceFilters, diags := toTerraformWorkspaceFilters(ctx, driftDetection.WorkspaceFilters)
	resp.Diagnostics.Append(diags...)
	plan.WorkspaceFilters = tfWorkspaceFilters

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *driftDetectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state driftDetectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	driftDetection, err := r.Client.DriftDetections.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving scalr_drift_detection", err.Error())
		return
	}

	state.EnvironmentID = types.StringValue(driftDetection.Environment.ID)
	state.CheckPeriod = types.StringValue(string(driftDetection.Schedule))
	state.RunMode = types.StringValue(string(driftDetection.RunMode))
	tfWorkspaceFilters, diags := toTerraformWorkspaceFilters(ctx, driftDetection.WorkspaceFilters)
	resp.Diagnostics.Append(diags...)
	state.WorkspaceFilters = tfWorkspaceFilters

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *driftDetectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan driftDetectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceFilter, diags := toScalrWorkspaceFilters(ctx, plan.WorkspaceFilters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.DriftDetectionUpdateOptions{
		Environment:      &scalr.Environment{ID: plan.EnvironmentID.ValueString()},
		Schedule:         scalr.DriftDetectionSchedulePeriod(plan.CheckPeriod.ValueString()),
		WorkspaceFilters: *workspaceFilter,
	}

	if !plan.RunMode.IsUnknown() && !plan.RunMode.IsNull() {
		opts.RunMode = ptr(scalr.DriftDetectionScheduleRunMode(plan.RunMode.ValueString()))
	}

	driftDetection, err := r.Client.DriftDetections.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating scalr_drift_detection", err.Error())
		return
	}

	plan.CheckPeriod = types.StringValue(string(driftDetection.Schedule))
	plan.RunMode = types.StringValue(string(driftDetection.RunMode))

	tfWorkspaceFilters, diags := toTerraformWorkspaceFilters(ctx, driftDetection.WorkspaceFilters)
	resp.Diagnostics.Append(diags...)
	plan.WorkspaceFilters = tfWorkspaceFilters

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *driftDetectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state driftDetectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.DriftDetections.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting scalr_drift_detection", err.Error())
		return
	}
}

func (r *driftDetectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *driftDetectionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var cfgSet, planSet types.Set

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("workspace_filters"), &cfgSet)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("workspace_filters"), &planSet)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !cfgSet.IsNull() && !cfgSet.IsUnknown() {
		var cfgFilters []workspaceFiltersModel
		resp.Diagnostics.Append(cfgSet.ElementsAs(ctx, &cfgFilters, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(cfgFilters) > 0 {
			return
		}
	}

	if !planSet.IsNull() && !planSet.IsUnknown() {
		var planFilters []workspaceFiltersModel
		resp.Diagnostics.Append(planSet.ElementsAs(ctx, &planFilters, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(planFilters) > 0 {
			return
		}
	}

	filters := []workspaceFiltersModel{
		{
			NamePatterns:     types.SetNull(types.StringType),
			EnvironmentTypes: types.SetNull(types.StringType),
			Tags:             types.SetNull(types.StringType),
		},
	}

	tfSet, d := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name_patterns":     types.SetType{ElemType: types.StringType},
			"environment_types": types.SetType{ElemType: types.StringType},
			"tags":              types.SetType{ElemType: types.StringType},
		},
	}, filters)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(
		resp.Plan.SetAttribute(ctx, path.Root("workspace_filters"), tfSet)...,
	)
}
