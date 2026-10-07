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

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ resource.Resource                = &datadogIntegrationResource{}
	_ resource.ResourceWithConfigure   = &datadogIntegrationResource{}
	_ resource.ResourceWithImportState = &datadogIntegrationResource{}
)

func newDatadogIntegrationResource() resource.Resource {
	return &datadogIntegrationResource{}
}

// datadogIntegrationResource defines the resource implementation.
type datadogIntegrationResource struct {
	framework.ResourceWithScalrClient
}

// datadogIntegrationResourceModel describes the resource data model.
type datadogIntegrationResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ApiKey        types.String `tfsdk:"api_key"`
	DeploymentUrl types.String `tfsdk:"deployment_url"`
	Status        types.String `tfsdk:"status"`
	ErrMessage    types.String `tfsdk:"err_message"`
	AccountID     types.String `tfsdk:"account_id"`
}

// refresh overwrites the model with the values returned by the API.
// The API key is never returned, so the value from the plan or state is kept.
func (m *datadogIntegrationResourceModel) refresh(di *schemas.DatadogIntegration) {
	m.Id = types.StringValue(di.ID)
	m.Name = types.StringValue(di.Attributes.Name)
	m.DeploymentUrl = types.StringPointerValue(di.Attributes.DeploymentUrl)
	m.ErrMessage = types.StringPointerValue(di.Attributes.ErrMessage)
	m.Status = types.StringValue(string(di.Attributes.Status))
	if di.Relationships.Account != nil {
		m.AccountID = types.StringValue(di.Relationships.Account.ID)
	}
}

func (r *datadogIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datadog_integration"
}

func (r *datadogIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of Datadog integrations in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Datadog integration.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(128),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Datadog API key.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"deployment_url": schema.StringAttribute{
				MarkdownDescription: "URL of the Datadog site, e.g. `https://us5.datadoghq.com`. If not set, `https://api.datadoghq.com` is used.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration: `active`, `disabled` or `failed`. " +
					"Scalr sets the `failed` status when the integration cannot connect to Datadog.",
				Computed: true,
			},
			"err_message": schema.StringAttribute{
				MarkdownDescription: "Error message reported by Scalr when the integration is in the `failed` status.",
				Computed:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account the integration belongs to.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *datadogIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan datadogIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.DatadogIntegrationRequest{
		Attributes: schemas.DatadogIntegrationAttributesRequest{
			Name:          value.Set(plan.Name.ValueString()),
			ApiKey:        value.Set(plan.ApiKey.ValueString()),
			DeploymentUrl: value.SetPtrMaybe(plan.DeploymentUrl.ValueStringPointer()),
		},
	}

	di, err := r.ClientV2.DatadogIntegration.CreateDatadogIntegration(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Datadog integration", err.Error())
		return
	}

	plan.refresh(di)
	if plan.ErrMessage.ValueString() != "" {
		resp.Diagnostics.AddWarning("Datadog integration has failed", plan.ErrMessage.ValueString())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *datadogIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datadogIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	di, err := r.ClientV2.DatadogIntegration.GetDatadogIntegration(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Datadog integration", err.Error())
		return
	}

	state.refresh(di)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state datadogIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.DatadogIntegrationRequest{
		Attributes: schemas.DatadogIntegrationAttributesRequest{
			Name: value.Set(plan.Name.ValueString()),
		},
	}
	if !plan.ApiKey.Equal(state.ApiKey) {
		opts.Attributes.ApiKey = value.Set(plan.ApiKey.ValueString())
	}
	if !plan.DeploymentUrl.Equal(state.DeploymentUrl) {
		opts.Attributes.DeploymentUrl = value.SetPtr(plan.DeploymentUrl.ValueStringPointer())
	}

	di, err := r.ClientV2.DatadogIntegration.UpdateDatadogIntegrations(ctx, plan.Id.ValueString(), &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Datadog integration", err.Error())
		return
	}

	plan.refresh(di)
	if plan.ErrMessage.ValueString() != "" {
		resp.Diagnostics.AddWarning("Datadog integration has failed", plan.ErrMessage.ValueString())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *datadogIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datadogIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.DatadogIntegration.DeleteDatadogIntegration(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting Datadog integration", err.Error())
		return
	}
}

func (r *datadogIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
