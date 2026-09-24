package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

var (
	_ resource.Resource                = &moduleTestProviderConfigurationResource{}
	_ resource.ResourceWithConfigure   = &moduleTestProviderConfigurationResource{}
	_ resource.ResourceWithImportState = &moduleTestProviderConfigurationResource{}
)

func newModuleTestProviderConfigurationResource() resource.Resource {
	return &moduleTestProviderConfigurationResource{}
}

type moduleTestProviderConfigurationResource struct {
	framework.ResourceWithScalrClient
}

type moduleTestProviderConfigurationResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	TestConfigurationID     types.String `tfsdk:"test_configuration_id"`
	ProviderConfigurationID types.String `tfsdk:"provider_configuration_id"`
}

func (r *moduleTestProviderConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_test_provider_configuration"
}

func (r *moduleTestProviderConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a provider configuration (credentials) to a module test configuration, so it can be used while running the module's tests." +
			"\n\n-> **Note** The provider configuration must have `is_allowed_in_module_test = true` (see [`scalr_provider_configuration`](provider_resource_scalr_provider_configuration))." +
			" Only one provider configuration per provider type can be linked to a given test configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"test_configuration_id": schema.StringAttribute{
				MarkdownDescription: "ID of the module test configuration, in the format `tc-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"provider_configuration_id": schema.StringAttribute{
				MarkdownDescription: "ID of the provider configuration to use as credentials for the module tests, in the format `pcfg-<RANDOM STRING>`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
		},
	}
}

func moduleTestProviderConfigurationModelFromAPI(link *scalr.ModuleTestProviderConfigurationLink) moduleTestProviderConfigurationResourceModel {
	model := moduleTestProviderConfigurationResourceModel{
		ID: types.StringValue(link.ID),
	}
	if link.TestConfiguration != nil {
		model.TestConfigurationID = types.StringValue(link.TestConfiguration.ID)
	}
	if link.ProviderConfiguration != nil {
		model.ProviderConfigurationID = types.StringValue(link.ProviderConfiguration.ID)
	}
	return model
}

func (r *moduleTestProviderConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan moduleTestProviderConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.ModuleTestProviderConfigurationLinkCreateOptions{
		ProviderConfiguration: &scalr.ProviderConfiguration{ID: plan.ProviderConfigurationID.ValueString()},
	}

	link, err := r.Client.ModuleTestProviderConfigurationLinks.Create(ctx, plan.TestConfigurationID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating module test provider configuration", err.Error())
		return
	}

	result := moduleTestProviderConfigurationModelFromAPI(link)
	if result.TestConfigurationID.ValueString() == "" {
		result.TestConfigurationID = plan.TestConfigurationID
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *moduleTestProviderConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state moduleTestProviderConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	link, err := r.Client.ModuleTestProviderConfigurationLinks.Read(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving module test provider configuration", err.Error())
		return
	}

	result := moduleTestProviderConfigurationModelFromAPI(link)
	if result.TestConfigurationID.ValueString() == "" {
		result.TestConfigurationID = state.TestConfigurationID
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *moduleTestProviderConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan moduleTestProviderConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.ModuleTestProviderConfigurationLinkUpdateOptions{
		ProviderConfiguration: &scalr.ProviderConfiguration{ID: plan.ProviderConfigurationID.ValueString()},
	}

	link, err := r.Client.ModuleTestProviderConfigurationLinks.Update(ctx, plan.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating module test provider configuration", err.Error())
		return
	}

	result := moduleTestProviderConfigurationModelFromAPI(link)
	if result.TestConfigurationID.ValueString() == "" {
		result.TestConfigurationID = plan.TestConfigurationID
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *moduleTestProviderConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state moduleTestProviderConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.ModuleTestProviderConfigurationLinks.Delete(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting module test provider configuration", err.Error())
		return
	}
}

func (r *moduleTestProviderConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
