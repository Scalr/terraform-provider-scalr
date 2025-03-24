package provider

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/scalr/go-scalr"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &integrationInfracostResource{}
	_ resource.ResourceWithConfigure        = &integrationInfracostResource{}
	_ resource.ResourceWithConfigValidators = &integrationInfracostResource{}
	_ resource.ResourceWithImportState      = &integrationInfracostResource{}
)

func newIntegrationInfracostResource() resource.Resource {
	return &integrationInfracostResource{}
}

// integrationInfracostResource defines the resource implementation.
type integrationInfracostResource struct {
	framework.ResourceWithScalrClient
}

// integrationInfracostResourceModel describes the resource data model.
type integrationInfracostResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ApiKey       types.String `tfsdk:"api_key"`
	Environments types.Set    `tfsdk:"environments"`
}

func (r *integrationInfracostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_infracost"
}

func (r *integrationInfracostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of Infracost integration in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Infracost integration.",
				Required:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for the Infracost integration.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of environments this integration is linked to. Use `[\"*\"]` to allow in all environments.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *integrationInfracostResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *integrationInfracostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan integrationInfracostResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.InfracostIntegrationCreateOptions{
		Name:     plan.Name.ValueStringPointer(),
		ApiKey:   plan.ApiKey.ValueStringPointer(),
		IsShared: ptr(false),
	}

	if !plan.Environments.IsUnknown() && !plan.Environments.IsNull() {
		var environments []string
		resp.Diagnostics.Append(plan.Environments.ElementsAs(ctx, &environments, false)...)

		if (len(environments) == 1) && (environments[0] == "*") {
			opts.IsShared = ptr(true)
		} else if len(environments) > 0 {
			envs := make([]*scalr.Environment, len(environments))
			for i, env := range environments {
				envs[i] = &scalr.Environment{ID: env}
			}
			opts.Environments = envs
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	integrationInfracost, err := r.Client.InfracostIntegrations.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Infracost integration", err.Error())
		return
	}

	plan.Id = types.StringValue(integrationInfracost.ID)
	plan.Name = types.StringValue(integrationInfracost.Name)

	envs := make([]string, len(integrationInfracost.Environments))
	for i, env := range integrationInfracost.Environments {
		envs[i] = env.ID
	}
	if integrationInfracost.IsShared {
		envs = []string{"*"}
	}
	envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
	resp.Diagnostics.Append(d...)
	plan.Environments = envsValues

	if len(integrationInfracost.ErrMessage) > 0 {
		resp.Diagnostics.AddWarning("Issues detected", integrationInfracost.ErrMessage)
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *integrationInfracostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state integrationInfracostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	integrationInfracost, err := r.Client.InfracostIntegrations.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Infracost integration", err.Error())
		return
	}

	// Overwrite attributes with refreshed values
	state.Name = types.StringValue(integrationInfracost.Name)

	if integrationInfracost.IsShared {
		envs := []string{"*"}
		envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
		resp.Diagnostics.Append(d...)
		state.Environments = envsValues
	} else {
		envs := make([]string, len(integrationInfracost.Environments))
		for i, env := range integrationInfracost.Environments {
			envs[i] = env.ID
		}
		envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
		resp.Diagnostics.Append(d...)
		state.Environments = envsValues
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *integrationInfracostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state integrationInfracostResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.InfracostIntegrationUpdateOptions{}

	if !plan.Name.Equal(state.Name) {
		opts.Name = plan.Name.ValueStringPointer()
	}

	if !plan.ApiKey.Equal(state.ApiKey) {
		opts.ApiKey = plan.ApiKey.ValueStringPointer()
	}

	var environments []string
	if !plan.Environments.Equal(state.Environments) {
		resp.Diagnostics.Append(plan.Environments.ElementsAs(ctx, &environments, false)...)
	} else {
		resp.Diagnostics.Append(state.Environments.ElementsAs(ctx, &environments, false)...)
	}
	if (len(environments) == 1) && (environments[0] == "*") {
		opts.IsShared = ptr(true)
		opts.Environments = make([]*scalr.Environment, 0)
	} else if len(environments) > 0 {
		envs := make([]*scalr.Environment, len(environments))
		for i, env := range environments {
			envs[i] = &scalr.Environment{ID: env}
		}
		opts.Environments = envs
		opts.IsShared = ptr(false)
	} else {
		opts.IsShared = ptr(false)
		opts.Environments = make([]*scalr.Environment, 0)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing resource
	integrationInfracost, err := r.Client.InfracostIntegrations.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Infracost integration", err.Error())
		return
	}

	// Overwrite attributes with refreshed values
	plan.Name = types.StringValue(integrationInfracost.Name)

	if integrationInfracost.IsShared {
		envs := []string{"*"}
		envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
		resp.Diagnostics.Append(d...)
		state.Environments = envsValues
	} else {
		envs := make([]string, len(integrationInfracost.Environments))
		for i, env := range integrationInfracost.Environments {
			envs[i] = env.ID
		}
		envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
		resp.Diagnostics.Append(d...)
		state.Environments = envsValues
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *integrationInfracostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state integrationInfracostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.InfracostIntegrations.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting Infracost integration", err.Error())
		return
	}
}

func (r *integrationInfracostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
