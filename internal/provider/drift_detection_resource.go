package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
)

func newDriftDetectionResource() resource.Resource {
	return &driftDetectionResource{}
}

type driftDetectionResource struct {
	framework.ResourceWithScalrClient
}

type driftDetectionResourceModel struct {
	Id            types.String `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	CheckPeriod   types.String `tfsdk:"check_period"`
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
		},
	}
}

func (r *driftDetectionResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *driftDetectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan driftDetectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.DriftDetectionCreateOptions{
		Environment: &scalr.Environment{ID: plan.EnvironmentID.ValueString()},
		Schedule:    scalr.DriftDetectionSchedulePeriod(plan.CheckPeriod.ValueString()),
	}
	driftDetection, err := r.Client.DriftDetections.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating scalr_drift_detection", err.Error())
		return
	}

	plan.Id = types.StringValue(driftDetection.ID)
	plan.EnvironmentID = types.StringValue(driftDetection.Environment.ID)
	plan.CheckPeriod = types.StringValue(string(driftDetection.Schedule))

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

	opts := scalr.DriftDetectionUpdateOptions{
		Environment: &scalr.Environment{ID: plan.EnvironmentID.ValueString()},
		Schedule:    scalr.DriftDetectionSchedulePeriod(plan.CheckPeriod.ValueString()),
	}

	driftDetection, err := r.Client.DriftDetections.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating scalr_drift_detection", err.Error())
		return
	}

	plan.CheckPeriod = types.StringValue(string(driftDetection.Schedule))

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
