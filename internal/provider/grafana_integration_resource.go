package provider

import (
	"context"
	"errors"
	"net/url"
	"regexp"
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

// grafanaLogsPath is the OTLP logs path Scalr appends to the configured endpoint.
const grafanaLogsPath = "/v1/logs"

var grafanaTenantIDRegexp = regexp.MustCompile(`^[^:]*$`)

// Compile-time interface checks
var (
	_ resource.Resource                   = &grafanaIntegrationResource{}
	_ resource.ResourceWithConfigure      = &grafanaIntegrationResource{}
	_ resource.ResourceWithValidateConfig = &grafanaIntegrationResource{}
	_ resource.ResourceWithImportState    = &grafanaIntegrationResource{}
)

func newGrafanaIntegrationResource() resource.Resource {
	return &grafanaIntegrationResource{}
}

// grafanaIntegrationResource defines the resource implementation.
type grafanaIntegrationResource struct {
	framework.ResourceWithScalrClient
}

// grafanaIntegrationResourceModel describes the resource data model.
type grafanaIntegrationResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Kind       types.String `tfsdk:"kind"`
	Endpoint   types.String `tfsdk:"endpoint"`
	AuthType   types.String `tfsdk:"auth_type"`
	Username   types.String `tfsdk:"username"`
	Secret     types.String `tfsdk:"secret"`
	TenantID   types.String `tfsdk:"tenant_id"`
	Status     types.String `tfsdk:"status"`
	ErrMessage types.String `tfsdk:"err_message"`
	AccountID  types.String `tfsdk:"account_id"`
}

// refresh overwrites the model with the values returned by the API.
// The secret is never returned, so the value from the plan or state is kept.
func (m *grafanaIntegrationResourceModel) refresh(gi *schemas.GrafanaIntegration) {
	m.Id = types.StringValue(gi.ID)
	m.Name = types.StringValue(gi.Attributes.Name)
	m.Kind = types.StringValue(string(gi.Attributes.Kind))
	m.Endpoint = types.StringValue(gi.Attributes.Endpoint)
	m.AuthType = types.StringValue(string(gi.Attributes.AuthType))
	m.Username = types.StringPointerValue(gi.Attributes.Username)
	m.TenantID = types.StringPointerValue(gi.Attributes.TenantId)
	m.Status = types.StringValue(string(gi.Attributes.Status))
	m.ErrMessage = types.StringPointerValue(gi.Attributes.ErrMessage)
	if gi.Relationships.Account != nil {
		m.AccountID = types.StringValue(gi.Relationships.Account.ID)
	}
}

// request builds the API payload from the model. All settings are sent on update too,
// since the API validates them together.
func (m *grafanaIntegrationResourceModel) request() *schemas.GrafanaIntegrationRequest {
	return &schemas.GrafanaIntegrationRequest{
		Attributes: schemas.GrafanaIntegrationAttributesRequest{
			Name:     value.Set(m.Name.ValueString()),
			Kind:     value.Set(schemas.GrafanaIntegrationKind(m.Kind.ValueString())),
			Endpoint: value.Set(m.Endpoint.ValueString()),
			AuthType: value.Set(schemas.GrafanaIntegrationAuthType(m.AuthType.ValueString())),
			Username: value.SetPtr(m.Username.ValueStringPointer()),
			TenantId: value.SetPtr(m.TenantID.ValueStringPointer()),
			Secret:   value.SetPtrMaybe(m.Secret.ValueStringPointer()),
		},
	}
}

func (r *grafanaIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grafana_integration"
}

func (r *grafanaIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of Grafana integrations in Scalr." +
			" The integration sends Scalr audit logs to Grafana Cloud or a self-hosted Loki over OTLP." +
			" Scalr checks the connection on create and update; when the check is rejected," +
			" the integration is still saved with the `failed` status.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Grafana integration.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(128),
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Grafana destination: `cloud` for Grafana Cloud or `loki` for a self-hosted Loki.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(schemas.GrafanaIntegrationKindCloud),
						string(schemas.GrafanaIntegrationKindLoki),
					),
				},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base OTLP URL, for example `https://otlp-gateway-prod-eu-west-2.grafana.net/otlp`." +
					" Must use https and must not end with `/v1/logs`.",
				Required: true,
				Validators: []validator.String{
					stringvalidation.StringIsValidURL(),
				},
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "Authentication scheme: `none`, `basic` or `bearer`. Grafana Cloud requires `basic`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(schemas.GrafanaIntegrationAuthTypeNone),
						string(schemas.GrafanaIntegrationAuthTypeBasic),
						string(schemas.GrafanaIntegrationAuthTypeBearer),
					),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Basic auth username. On Grafana Cloud, the numeric instance ID." +
					" Required for `basic` auth, not used with other auth types.",
				Optional: true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "Basic auth password or bearer token. Required for `basic` and `bearer` auth." +
					" Scalr never returns this value, so it is kept in the state as configured.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "Loki tenant ID, sent as the `X-Scope-OrgID` header. For a multi-tenant `loki` only." +
					" Must not contain `:`.",
				Optional: true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.RegexMatches(grafanaTenantIDRegexp, "must not contain ':'"),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration: `active`, `disabled`, `failed` or `pending`." +
					" Scalr sets the `failed` status when the connection check is rejected.",
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

func (r *grafanaIntegrationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg grafanaIntegrationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !cfg.Endpoint.IsNull() && !cfg.Endpoint.IsUnknown() {
		if u, err := url.Parse(cfg.Endpoint.ValueString()); err == nil &&
			strings.HasSuffix(strings.TrimRight(u.Path, "/"), grafanaLogsPath) {
			resp.Diagnostics.AddAttributeError(
				path.Root("endpoint"),
				"Invalid endpoint",
				"Enter the base OTLP URL without "+grafanaLogsPath+".",
			)
		}
	}

	kind := cfg.Kind
	authType := cfg.AuthType

	if !kind.IsUnknown() && kind.ValueString() == string(schemas.GrafanaIntegrationKindCloud) {
		if !authType.IsUnknown() && !authType.IsNull() &&
			authType.ValueString() != string(schemas.GrafanaIntegrationAuthTypeBasic) {
			resp.Diagnostics.AddAttributeError(
				path.Root("auth_type"),
				"Invalid auth type",
				"Grafana Cloud requires basic auth.",
			)
		}
		if !cfg.TenantID.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("tenant_id"),
				"Invalid attribute combination",
				"The argument \"tenant_id\" is only used with kind \"loki\".",
			)
		}
	}

	if authType.IsUnknown() || authType.IsNull() {
		return
	}

	switch schemas.GrafanaIntegrationAuthType(authType.ValueString()) {
	case schemas.GrafanaIntegrationAuthTypeBasic:
		requireGrafanaAttribute(resp, "username", cfg.Username, authType.ValueString())
		requireGrafanaAttribute(resp, "secret", cfg.Secret, authType.ValueString())
	case schemas.GrafanaIntegrationAuthTypeBearer:
		rejectGrafanaAttribute(resp, "username", cfg.Username, authType.ValueString())
		requireGrafanaAttribute(resp, "secret", cfg.Secret, authType.ValueString())
	case schemas.GrafanaIntegrationAuthTypeNone:
		rejectGrafanaAttribute(resp, "username", cfg.Username, authType.ValueString())
		rejectGrafanaAttribute(resp, "secret", cfg.Secret, authType.ValueString())
	}
}

func requireGrafanaAttribute(resp *resource.ValidateConfigResponse, name string, v types.String, authType string) {
	if v.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root(name),
			"Missing required argument",
			"The argument \""+name+"\" is required when \"auth_type\" is \""+authType+"\".",
		)
	}
}

func rejectGrafanaAttribute(resp *resource.ValidateConfigResponse, name string, v types.String, authType string) {
	if !v.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root(name),
			"Invalid attribute combination",
			"The argument \""+name+"\" is not used when \"auth_type\" is \""+authType+"\".",
		)
	}
}

func (r *grafanaIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan grafanaIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gi, err := r.ClientV2.GrafanaIntegration.CreateGrafanaIntegration(ctx, plan.request())
	if err != nil {
		resp.Diagnostics.AddError("Error creating Grafana integration", err.Error())
		return
	}

	plan.refresh(gi)
	if plan.ErrMessage.ValueString() != "" {
		resp.Diagnostics.AddWarning("Grafana integration has failed", plan.ErrMessage.ValueString())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *grafanaIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state grafanaIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gi, err := r.ClientV2.GrafanaIntegration.GetGrafanaIntegration(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Grafana integration", err.Error())
		return
	}

	state.refresh(gi)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *grafanaIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan grafanaIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gi, err := r.ClientV2.GrafanaIntegration.UpdateGrafanaIntegration(ctx, plan.Id.ValueString(), plan.request())
	if err != nil {
		resp.Diagnostics.AddError("Error updating Grafana integration", err.Error())
		return
	}

	plan.refresh(gi)
	if plan.ErrMessage.ValueString() != "" {
		resp.Diagnostics.AddWarning("Grafana integration has failed", plan.ErrMessage.ValueString())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *grafanaIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state grafanaIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.GrafanaIntegration.DeleteGrafanaIntegration(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting Grafana integration", err.Error())
		return
	}
}

func (r *grafanaIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
