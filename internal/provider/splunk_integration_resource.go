package provider

import (
	"context"
	"errors"
	"net/url"
	"strings"

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

// splunkCollectorPath is the HEC path prefix; Scalr appends the event endpoint to the configured URL.
const splunkCollectorPath = "/services/collector"

// Compile-time interface checks
var (
	_ resource.Resource                   = &splunkIntegrationResource{}
	_ resource.ResourceWithConfigure      = &splunkIntegrationResource{}
	_ resource.ResourceWithValidateConfig = &splunkIntegrationResource{}
	_ resource.ResourceWithImportState    = &splunkIntegrationResource{}
)

func newSplunkIntegrationResource() resource.Resource {
	return &splunkIntegrationResource{}
}

// splunkIntegrationResource defines the resource implementation.
type splunkIntegrationResource struct {
	framework.ResourceWithScalrClient
}

// splunkIntegrationResourceModel describes the resource data model.
type splunkIntegrationResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Endpoint   types.String `tfsdk:"endpoint"`
	Token      types.String `tfsdk:"token"`
	Index      types.String `tfsdk:"index"`
	Source     types.String `tfsdk:"source"`
	Sourcetype types.String `tfsdk:"sourcetype"`
	Status     types.String `tfsdk:"status"`
	ErrMessage types.String `tfsdk:"err_message"`
	AccountID  types.String `tfsdk:"account_id"`
}

// The API never returns the token, so it is taken from the plan or state.
func splunkIntegrationResourceModelFromAPI(si *schemas.SplunkIntegration, token types.String) *splunkIntegrationResourceModel {
	model := &splunkIntegrationResourceModel{
		Id:         types.StringValue(si.ID),
		Name:       types.StringValue(si.Attributes.Name),
		Endpoint:   types.StringValue(si.Attributes.Endpoint),
		Token:      token,
		Index:      types.StringValue(si.Attributes.Index),
		Source:     types.StringPointerValue(si.Attributes.Source),
		Sourcetype: types.StringPointerValue(si.Attributes.Sourcetype),
		Status:     types.StringValue(string(si.Attributes.Status)),
		ErrMessage: types.StringPointerValue(si.Attributes.ErrMessage),
		AccountID:  types.StringNull(),
	}

	if si.Relationships.Account != nil {
		model.AccountID = types.StringValue(si.Relationships.Account.ID)
	}

	return model
}

func (r *splunkIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_splunk_integration"
}

func (r *splunkIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of Splunk integrations in Scalr." +
			" The integration sends Scalr audit logs and run events to a Splunk HTTP Event Collector (HEC).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Splunk integration.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(128),
				},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base HEC URL, for example `https://http-inputs-<stack>.splunkcloud.com` or `https://splunk.example.com:8088`." +
					" Must use https and must not include `/services/collector`.",
				Required: true,
				Validators: []validator.String{
					stringvalidation.StringIsValidURL(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "HEC token. Indexer acknowledgment must be turned off for it.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"index": schema.StringAttribute{
				MarkdownDescription: "Splunk index for the events. The HEC token must be allowed to write to it.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(255),
				},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "Labels where events come from. Use it to filter Scalr events in Splunk searches." +
					" Scalr sends `scalr` when it is not set.",
				Optional: true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(255),
				},
			},
			"sourcetype": schema.StringAttribute{
				MarkdownDescription: "Tells Splunk how to parse events. Scalr sends JSON, so keep the default `_json`" +
					" unless your Splunk admin uses a custom sourcetype. Scalr sends `_json` when it is not set.",
				Optional: true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(255),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration: `active`, `disabled` or `failed`." +
					" Scalr sets the `failed` status when Splunk rejects the token or the index during delivery.",
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

func (r *splunkIntegrationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg splunkIntegrationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.Endpoint.IsNull() || cfg.Endpoint.IsUnknown() {
		return
	}
	if u, err := url.Parse(cfg.Endpoint.ValueString()); err == nil && strings.Contains(u.Path, splunkCollectorPath) {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Invalid endpoint",
			"Enter the base HEC URL without "+splunkCollectorPath+".",
		)
	}
}

func (r *splunkIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan splunkIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.SplunkIntegrationRequest{
		Attributes: schemas.SplunkIntegrationAttributesRequest{
			Name:       value.Set(plan.Name.ValueString()),
			Endpoint:   value.Set(plan.Endpoint.ValueString()),
			Token:      value.Set(plan.Token.ValueString()),
			Index:      value.Set(plan.Index.ValueString()),
			Source:     value.SetPtr(plan.Source.ValueStringPointer()),
			Sourcetype: value.SetPtr(plan.Sourcetype.ValueStringPointer()),
		},
	}

	si, err := r.ClientV2.SplunkIntegration.CreateSplunkIntegration(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Splunk integration", err.Error())
		return
	}

	result := splunkIntegrationResourceModelFromAPI(si, plan.Token)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *splunkIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state splunkIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	si, err := r.ClientV2.SplunkIntegration.GetSplunkIntegration(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Splunk integration", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, splunkIntegrationResourceModelFromAPI(si, state.Token))...)
}

func (r *splunkIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state splunkIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.SplunkIntegrationRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.Endpoint.Equal(state.Endpoint) {
		opts.Attributes.Endpoint = value.Set(plan.Endpoint.ValueString())
	}

	// The API drops the stored token when the endpoint changes and no token is sent,
	// so send it along with the endpoint.
	if !plan.Token.Equal(state.Token) || !plan.Endpoint.Equal(state.Endpoint) {
		opts.Attributes.Token = value.Set(plan.Token.ValueString())
	}

	if !plan.Index.Equal(state.Index) {
		opts.Attributes.Index = value.Set(plan.Index.ValueString())
	}

	if !plan.Source.Equal(state.Source) {
		opts.Attributes.Source = value.SetPtr(plan.Source.ValueStringPointer())
	}

	if !plan.Sourcetype.Equal(state.Sourcetype) {
		opts.Attributes.Sourcetype = value.SetPtr(plan.Sourcetype.ValueStringPointer())
	}

	si, err := r.ClientV2.SplunkIntegration.UpdateSplunkIntegration(ctx, plan.Id.ValueString(), &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Splunk integration", err.Error())
		return
	}

	result := splunkIntegrationResourceModelFromAPI(si, plan.Token)
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *splunkIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state splunkIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.SplunkIntegration.DeleteSplunkIntegration(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting Splunk integration", err.Error())
		return
	}
}

func (r *splunkIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
