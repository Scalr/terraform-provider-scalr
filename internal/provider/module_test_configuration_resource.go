package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

var (
	_ resource.Resource                = &moduleTestConfigurationResource{}
	_ resource.ResourceWithConfigure   = &moduleTestConfigurationResource{}
	_ resource.ResourceWithImportState = &moduleTestConfigurationResource{}
)

func newModuleTestConfigurationResource() resource.Resource {
	return &moduleTestConfigurationResource{}
}

type moduleTestConfigurationResource struct {
	framework.ResourceWithScalrClient
}

type moduleTestConfigurationResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	ModuleID                   types.String `tfsdk:"module_id"`
	Enabled                    types.Bool   `tfsdk:"enabled"`
	FailureBehavior            types.String `tfsdk:"failure_behavior"`
	TriggerOnPrActivityEnabled types.Bool   `tfsdk:"trigger_on_pr_activity_enabled"`
	TriggerOnNewVersionEnabled types.Bool   `tfsdk:"trigger_on_new_version_enabled"`
}

func (r *moduleTestConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_test_configuration"
}

func (r *moduleTestConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the tofu test configuration of a module in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"module_id": schema.StringAttribute{
				MarkdownDescription: "ID of the module, in the format `mod-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the test configuration is enabled. Default: `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"failure_behavior": schema.StringAttribute{
				MarkdownDescription: "The behavior to apply when a test run fails. Valid values are `failure` and `notify`. Default: `notify`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(scalr.ModuleTestFailureBehaviorNotify)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.ModuleTestFailureBehaviorFailure),
						string(scalr.ModuleTestFailureBehaviorNotify),
					),
				},
			},
			"trigger_on_pr_activity_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether a test run should be triggered on pull request activity. Default: `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"trigger_on_new_version_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether a test run should be triggered when a new module version is synced. Default: `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func moduleTestConfigurationOptionsFromModel(model moduleTestConfigurationResourceModel) scalr.ModuleTestConfigurationUpdateOptions {
	failureBehavior := scalr.ModuleTestFailureBehavior(model.FailureBehavior.ValueString())

	return scalr.ModuleTestConfigurationUpdateOptions{
		Enabled:                    model.Enabled.ValueBoolPointer(),
		FailureBehavior:            &failureBehavior,
		TriggerOnPrActivityEnabled: model.TriggerOnPrActivityEnabled.ValueBoolPointer(),
		TriggerOnNewVersionEnabled: model.TriggerOnNewVersionEnabled.ValueBoolPointer(),
	}
}

func moduleTestConfigurationModelFromAPI(moduleID string, tc *scalr.ModuleTestConfiguration) moduleTestConfigurationResourceModel {
	model := moduleTestConfigurationResourceModel{
		ID:                         types.StringValue(tc.ID),
		ModuleID:                   types.StringValue(moduleID),
		Enabled:                    types.BoolValue(tc.Enabled),
		FailureBehavior:            types.StringValue(string(tc.FailureBehavior)),
		TriggerOnPrActivityEnabled: types.BoolValue(tc.TriggerOnPrActivityEnabled),
		TriggerOnNewVersionEnabled: types.BoolValue(tc.TriggerOnNewVersionEnabled),
	}
	if tc.Module != nil {
		model.ModuleID = types.StringValue(tc.Module.ID)
	}
	return model
}

func (r *moduleTestConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan moduleTestConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tc, err := r.Client.ModuleTestConfigurations.Update(
		ctx, plan.ModuleID.ValueString(), moduleTestConfigurationOptionsFromModel(plan),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating module test configuration", err.Error())
		return
	}

	result := moduleTestConfigurationModelFromAPI(plan.ModuleID.ValueString(), tc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *moduleTestConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state moduleTestConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tc *scalr.ModuleTestConfiguration
	var err error
	if state.ID.ValueString() != "" {
		tc, err = r.Client.ModuleTestConfigurations.Read(ctx, state.ID.ValueString())
	} else {
		// Imported by module_id: the test configuration ID isn't known yet, so
		// resolve (or lazily create, matching the API's get-or-create semantics) it via its module.
		tc, err = r.Client.ModuleTestConfigurations.Update(
			ctx, state.ModuleID.ValueString(), scalr.ModuleTestConfigurationUpdateOptions{},
		)
	}
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving module test configuration", err.Error())
		return
	}

	result := moduleTestConfigurationModelFromAPI(state.ModuleID.ValueString(), tc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *moduleTestConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan moduleTestConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tc, err := r.Client.ModuleTestConfigurations.Update(
		ctx, plan.ModuleID.ValueString(), moduleTestConfigurationOptionsFromModel(plan),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating module test configuration", err.Error())
		return
	}

	result := moduleTestConfigurationModelFromAPI(plan.ModuleID.ValueString(), tc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *moduleTestConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state moduleTestConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// There is no delete endpoint for a module test configuration: disable it instead.
	_, err := r.Client.ModuleTestConfigurations.Update(
		ctx, state.ModuleID.ValueString(), scalr.ModuleTestConfigurationUpdateOptions{Enabled: scalr.Bool(false)},
	)
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error disabling module test configuration", err.Error())
		return
	}
}

func (r *moduleTestConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("module_id"), req, resp)
}
