package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
	_ resource.Resource                     = &workloadIdentityProviderResource{}
	_ resource.ResourceWithConfigure        = &workloadIdentityProviderResource{}
	_ resource.ResourceWithConfigValidators = &workloadIdentityProviderResource{}
	_ resource.ResourceWithImportState      = &workloadIdentityProviderResource{}
)

func newWorkloadIdentityProviderResource() resource.Resource {
	return &workloadIdentityProviderResource{}
}

type workloadIdentityProviderResource struct {
	framework.ResourceWithScalrClient
}

type workloadIdentityProviderResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	URL              types.String `tfsdk:"url"`
	AllowedAudiences types.Set    `tfsdk:"allowed_audiences"`
}

func (r *workloadIdentityProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workload_identity_provider"
}

func (r *workloadIdentityProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of workload identity providers in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the workload identity provider.",
				Required:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL of the workload identity provider.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.StringIsValidURL(),
				},
			},
			"allowed_audiences": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 10),
				},
				MarkdownDescription: "Set of allowed audiences for the workload identity provider. Must contain at least 1 and at most 10 elements.",
			},
		},
	}
}

func (r *workloadIdentityProviderResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *workloadIdentityProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workloadIdentityProviderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allowedAudiences []string
	resp.Diagnostics.Append(plan.AllowedAudiences.ElementsAs(ctx, &allowedAudiences, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.WorkloadIdentityProviderCreateOptions{
		Name:             plan.Name.ValueStringPointer(),
		URL:              plan.URL.ValueStringPointer(),
		AllowedAudiences: allowedAudiences,
	}

	provider, err := r.Client.WorkloadIdentityProviders.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workload identity provider", err.Error())
		return
	}

	plan.ID = types.StringValue(provider.ID)
	plan.Name = types.StringValue(provider.Name)
	plan.URL = types.StringValue(provider.URL)

	allowedAudiencesSet, diags := types.SetValueFrom(ctx, types.StringType, provider.AllowedAudiences)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AllowedAudiences = allowedAudiencesSet

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workloadIdentityProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workloadIdentityProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider, err := r.Client.WorkloadIdentityProviders.Read(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving workload identity provider", err.Error())
		return
	}

	state.Name = types.StringValue(provider.Name)
	state.URL = types.StringValue(provider.URL)

	allowedAudiencesSet, diags := types.SetValueFrom(ctx, types.StringType, provider.AllowedAudiences)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AllowedAudiences = allowedAudiencesSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workloadIdentityProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workloadIdentityProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allowedAudiences []string
	resp.Diagnostics.Append(plan.AllowedAudiences.ElementsAs(ctx, &allowedAudiences, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.WorkloadIdentityProviderUpdateOptions{
		Name:             plan.Name.ValueStringPointer(),
		AllowedAudiences: allowedAudiences,
	}

	provider, err := r.Client.WorkloadIdentityProviders.Update(ctx, plan.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating workload identity provider", err.Error())
		return
	}

	plan.Name = types.StringValue(provider.Name)
	plan.URL = types.StringValue(provider.URL)
	allowedAudiencesSet, diags := types.SetValueFrom(ctx, types.StringType, provider.AllowedAudiences)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AllowedAudiences = allowedAudiencesSet

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workloadIdentityProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workloadIdentityProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.WorkloadIdentityProviders.Delete(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting workload identity provider", err.Error())
		return
	}
}

func (r *workloadIdentityProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
