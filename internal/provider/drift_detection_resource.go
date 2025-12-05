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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

var (
	_ resource.Resource                = &driftDetectionResource{}
	_ resource.ResourceWithConfigure   = &driftDetectionResource{}
	_ resource.ResourceWithImportState = &driftDetectionResource{}
)

var filtersAttrTypes = map[string]attr.Type{
	"name_patterns":     types.SetType{ElemType: types.StringType},
	"environment_types": types.SetType{ElemType: types.StringType},
	"tags":              types.SetType{ElemType: types.StringType},
}

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
	WorkspaceFilters types.Object `tfsdk:"workspace_filters"`
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
						string(schemas.DriftDetectionScheduleScheduleDaily),
						string(schemas.DriftDetectionScheduleScheduleWeekly),
					),
				},
			},
			"run_mode": schema.StringAttribute{
				MarkdownDescription: "Run mode for drift detection: `refresh-only` (default) or `plan`. ",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(schemas.DriftDetectionScheduleRunModeRefreshOnly),
						string(schemas.DriftDetectionScheduleRunModePlan),
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"workspace_filters": schema.SingleNestedBlock{
				MarkdownDescription: "Filters for workspaces to be included in drift detection. Only one type of filter can be specified: `name_patterns`, `environment_types` or `tags`.",
				Attributes: map[string]schema.Attribute{
					"name_patterns": schema.SetAttribute{
						MarkdownDescription: "Workspace name patterns to include in drift detection. Supports `*` wildcard (e.g., `prod-*`).",
						ElementType:         types.StringType,
						Optional:            true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
					"environment_types": schema.SetAttribute{
						MarkdownDescription: "Workspace environment types to include in drift detection. Allowed values: `production`, `staging`, `testing`, `development`, `unmapped`.",
						ElementType:         types.StringType,
						Optional:            true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(
									"production",
									"staging",
									"testing",
									"development",
									"unmapped",
								),
							),
						},
					},
					"tags": schema.SetAttribute{
						MarkdownDescription: "Workspace tag to include in drift detection. A workspace matches if it has at least one of the specified tags.",
						ElementType:         types.StringType,
						Optional:            true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
				},
				Validators: []validator.Object{
					validation.ExactlyOneOfIfObjectSet(
						path.MatchRelative().AtName("name_patterns"),
						path.MatchRelative().AtName("environment_types"),
						path.MatchRelative().AtName("tags"),
					),
				},
			},
		},
	}
}

func toWorkspaceFiltersRequest(ctx context.Context, filtersObj types.Object) (*schemas.DriftDetectionScheduleWorkspaceFiltersRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	var filters workspaceFiltersModel

	if filtersObj.IsNull() || filtersObj.IsUnknown() {
		return nil, diags
	}

	diags.Append(filtersObj.As(ctx, &filters, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	filtersRequest := &schemas.DriftDetectionScheduleWorkspaceFiltersRequest{}

	if !filters.NamePatterns.IsNull() && !filters.NamePatterns.IsUnknown() {
		var namePatterns []string
		diags.Append(filters.NamePatterns.ElementsAs(ctx, &namePatterns, false)...)
		if diags.HasError() {
			return nil, diags
		}
		filtersRequest.NamePatterns = value.Set(namePatterns)
	}
	if !filters.EnvironmentTypes.IsNull() && !filters.EnvironmentTypes.IsUnknown() {
		var environmentTypes []string
		diags.Append(filters.EnvironmentTypes.ElementsAs(ctx, &environmentTypes, false)...)
		if diags.HasError() {
			return nil, diags
		}
		filtersRequest.EnvironmentTypes = value.Set(environmentTypes)
	}
	if !filters.Tags.IsNull() && !filters.Tags.IsUnknown() {
		var tags []string
		diags.Append(filters.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return nil, diags
		}
		filtersRequest.Tags = value.Set(tags)
	}

	return filtersRequest, diags
}

func driftDetectionResourceModelFromAPI(ctx context.Context, driftDetection *schemas.DriftDetectionSchedule) (*driftDetectionResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &driftDetectionResourceModel{
		Id:               types.StringValue(driftDetection.ID),
		EnvironmentID:    types.StringValue(driftDetection.Relationships.Environment.ID),
		CheckPeriod:      types.StringValue(string(driftDetection.Attributes.Schedule)),
		RunMode:          types.StringValue(string(driftDetection.Attributes.RunMode)),
		WorkspaceFilters: types.ObjectNull(filtersAttrTypes),
	}

	filters := workspaceFiltersModel{
		NamePatterns:     types.SetNull(types.StringType),
		EnvironmentTypes: types.SetNull(types.StringType),
		Tags:             types.SetNull(types.StringType),
	}

	namePatterns, d := types.SetValueFrom(ctx, types.StringType, driftDetection.Attributes.WorkspaceFilters.NamePatterns)
	diags.Append(d...)
	filters.NamePatterns = namePatterns

	envTypes, d := types.SetValueFrom(ctx, types.StringType, driftDetection.Attributes.WorkspaceFilters.EnvironmentTypes)
	diags.Append(d...)
	filters.EnvironmentTypes = envTypes

	tags, d := types.SetValueFrom(ctx, types.StringType, driftDetection.Attributes.WorkspaceFilters.Tags)
	diags.Append(d...)
	filters.Tags = tags

	if !filters.NamePatterns.IsNull() || !filters.EnvironmentTypes.IsNull() || !filters.Tags.IsNull() {
		filtersValue, d := types.ObjectValueFrom(ctx, filtersAttrTypes, filters)
		diags.Append(d...)
		model.WorkspaceFilters = filtersValue
	}

	return model, diags
}

func (r *driftDetectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan driftDetectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filtersRequest, diags := toWorkspaceFiltersRequest(ctx, plan.WorkspaceFilters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.DriftDetectionScheduleRequest{
		Attributes: schemas.DriftDetectionScheduleAttributesRequest{
			RunMode:          value.Set(schemas.DriftDetectionScheduleRunMode(plan.RunMode.ValueString())),
			Schedule:         value.Set(schemas.DriftDetectionScheduleSchedule(plan.CheckPeriod.ValueString())),
			WorkspaceFilters: value.SetPtrMaybe(filtersRequest),
		},
		Relationships: schemas.DriftDetectionScheduleRelationshipsRequest{
			Environment: value.Set(schemas.Environment{ID: plan.EnvironmentID.ValueString()}),
		},
	}

	driftDetection, err := r.ClientV2.DriftDetectionSchedule.CreateDriftDetectionSchedule(ctx, &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating scalr_drift_detection", err.Error())
		return
	}

	result, diags := driftDetectionResourceModelFromAPI(ctx, driftDetection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
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

	driftDetection, err := r.ClientV2.DriftDetectionSchedule.GetDriftDetectionSchedule(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving scalr_drift_detection", err.Error())
		return
	}

	result, diags := driftDetectionResourceModelFromAPI(ctx, driftDetection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *driftDetectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state driftDetectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.DriftDetectionScheduleRequest{}

	if !plan.RunMode.Equal(state.RunMode) {
		opts.Attributes.RunMode = value.Set(schemas.DriftDetectionScheduleRunMode(plan.RunMode.ValueString()))
	}

	if !plan.CheckPeriod.Equal(state.CheckPeriod) {
		opts.Attributes.Schedule = value.Set(schemas.DriftDetectionScheduleSchedule(plan.CheckPeriod.ValueString()))
	}

	if !plan.WorkspaceFilters.Equal(state.WorkspaceFilters) {
		if plan.WorkspaceFilters.IsNull() {
			opts.Attributes.WorkspaceFilters = value.Null[schemas.DriftDetectionScheduleWorkspaceFiltersRequest]()
		} else {
			filtersRequest, diags := toWorkspaceFiltersRequest(ctx, plan.WorkspaceFilters)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			opts.Attributes.WorkspaceFilters = value.SetPtrMaybe(filtersRequest)
		}
	}

	driftDetection, err := r.ClientV2.DriftDetectionSchedule.UpdateDriftDetectionSchedule(ctx, plan.Id.ValueString(), &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating scalr_drift_detection", err.Error())
		return
	}

	result, diags := driftDetectionResourceModelFromAPI(ctx, driftDetection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
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

	err := r.ClientV2.DriftDetectionSchedule.DeleteDriftDetectionSchedule(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting scalr_drift_detection", err.Error())
		return
	}
}

func (r *driftDetectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
