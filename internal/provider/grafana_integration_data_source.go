package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	grafanaops "github.com/scalr/go-scalr/v2/scalr/ops/grafana_integration"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &grafanaIntegrationDataSource{}
	_ datasource.DataSourceWithConfigure        = &grafanaIntegrationDataSource{}
	_ datasource.DataSourceWithConfigValidators = &grafanaIntegrationDataSource{}
)

func newGrafanaIntegrationDataSource() datasource.DataSource {
	return &grafanaIntegrationDataSource{}
}

// grafanaIntegrationDataSource defines the data source implementation.
type grafanaIntegrationDataSource struct {
	framework.DataSourceWithScalrClient
}

// grafanaIntegrationDataSourceModel describes the data source data model.
type grafanaIntegrationDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Kind       types.String `tfsdk:"kind"`
	Endpoint   types.String `tfsdk:"endpoint"`
	AuthType   types.String `tfsdk:"auth_type"`
	Username   types.String `tfsdk:"username"`
	TenantID   types.String `tfsdk:"tenant_id"`
	Status     types.String `tfsdk:"status"`
	ErrMessage types.String `tfsdk:"err_message"`
	AccountID  types.String `tfsdk:"account_id"`
}

func (d *grafanaIntegrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grafana_integration"
}

func (d *grafanaIntegrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Grafana integration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Grafana integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Grafana integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Grafana destination: `cloud` or `loki`.",
				Computed:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base OTLP URL.",
				Computed:            true,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "Authentication scheme: `none`, `basic` or `bearer`.",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Basic auth username. On Grafana Cloud, the numeric instance ID.",
				Computed:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "Loki tenant ID, sent as the `X-Scope-OrgID` header.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration: `active`, `disabled`, `failed` or `pending`.",
				Computed:            true,
			},
			"err_message": schema.StringAttribute{
				MarkdownDescription: "Error message reported by Scalr when the integration is in the `failed` status.",
				Computed:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account the integration belongs to.",
				Computed:            true,
			},
		},
	}
}

func (d *grafanaIntegrationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *grafanaIntegrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg grafanaIntegrationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var gi *schemas.GrafanaIntegration
	var diags diag.Diagnostics
	if !cfg.Id.IsNull() {
		gi, diags = d.getByID(ctx, cfg)
	} else {
		gi, diags = d.findByName(ctx, cfg)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg.Id = types.StringValue(gi.ID)
	cfg.Name = types.StringValue(gi.Attributes.Name)
	cfg.Kind = types.StringValue(string(gi.Attributes.Kind))
	cfg.Endpoint = types.StringValue(gi.Attributes.Endpoint)
	cfg.AuthType = types.StringValue(string(gi.Attributes.AuthType))
	cfg.Username = types.StringPointerValue(gi.Attributes.Username)
	cfg.TenantID = types.StringPointerValue(gi.Attributes.TenantId)
	cfg.Status = types.StringValue(string(gi.Attributes.Status))
	cfg.ErrMessage = types.StringPointerValue(gi.Attributes.ErrMessage)
	if gi.Relationships.Account != nil {
		cfg.AccountID = types.StringValue(gi.Relationships.Account.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

func (d *grafanaIntegrationDataSource) notFound(cfg grafanaIntegrationDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Error retrieving Grafana integration",
		fmt.Sprintf("Could not find Grafana integration with ID '%s', name '%s'.", cfg.Id.ValueString(), cfg.Name.ValueString()),
	)
	return diags
}

func (d *grafanaIntegrationDataSource) getByID(ctx context.Context, cfg grafanaIntegrationDataSourceModel) (*schemas.GrafanaIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	gi, err := d.ClientV2.GrafanaIntegration.GetGrafanaIntegration(ctx, cfg.Id.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return nil, d.notFound(cfg)
		}
		diags.AddError("Error retrieving Grafana integration", err.Error())
		return nil, diags
	}

	if !cfg.Name.IsNull() && gi.Attributes.Name != cfg.Name.ValueString() {
		return nil, d.notFound(cfg)
	}

	return gi, diags
}

func (d *grafanaIntegrationDataSource) findByName(ctx context.Context, cfg grafanaIntegrationDataSourceModel) (*schemas.GrafanaIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	opts := grafanaops.ListGrafanaIntegrationsOptions{Filter: map[string]string{
		"name": cfg.Name.ValueString(),
	}}

	found, err := d.ClientV2.GrafanaIntegration.ListGrafanaIntegrations(ctx, &opts)
	if err != nil {
		diags.AddError("Error retrieving Grafana integration", err.Error())
		return nil, diags
	}

	switch len(found) {
	case 0:
		return nil, d.notFound(cfg)
	case 1:
		return found[0], diags
	default:
		diags.AddError(
			"Error retrieving Grafana integration",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return nil, diags
	}
}
