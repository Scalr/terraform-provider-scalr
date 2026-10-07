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
	ddops "github.com/scalr/go-scalr/v2/scalr/ops/datadog_integration"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &datadogIntegrationDataSource{}
	_ datasource.DataSourceWithConfigure        = &datadogIntegrationDataSource{}
	_ datasource.DataSourceWithConfigValidators = &datadogIntegrationDataSource{}
)

func newDatadogIntegrationDataSource() datasource.DataSource {
	return &datadogIntegrationDataSource{}
}

// datadogIntegrationDataSource defines the data source implementation.
type datadogIntegrationDataSource struct {
	framework.DataSourceWithScalrClient
}

// datadogIntegrationDataSourceModel describes the data source data model.
type datadogIntegrationDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	DeploymentUrl types.String `tfsdk:"deployment_url"`
	Status        types.String `tfsdk:"status"`
	ErrMessage    types.String `tfsdk:"err_message"`
	AccountID     types.String `tfsdk:"account_id"`
}

func (d *datadogIntegrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datadog_integration"
}

func (d *datadogIntegrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Datadog integration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Datadog integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Datadog integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"deployment_url": schema.StringAttribute{
				MarkdownDescription: "URL of the Datadog site.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration: `active`, `disabled` or `failed`.",
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

func (d *datadogIntegrationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *datadogIntegrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg datadogIntegrationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var di *schemas.DatadogIntegration
	var diags diag.Diagnostics
	if !cfg.Id.IsNull() {
		di, diags = d.getByID(ctx, cfg)
	} else {
		di, diags = d.findByName(ctx, cfg)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg.Id = types.StringValue(di.ID)
	cfg.Name = types.StringValue(di.Attributes.Name)
	cfg.DeploymentUrl = types.StringPointerValue(di.Attributes.DeploymentUrl)
	cfg.Status = types.StringValue(string(di.Attributes.Status))
	cfg.ErrMessage = types.StringPointerValue(di.Attributes.ErrMessage)
	if di.Relationships.Account != nil {
		cfg.AccountID = types.StringValue(di.Relationships.Account.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

func (d *datadogIntegrationDataSource) notFound(cfg datadogIntegrationDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Error retrieving Datadog integration",
		fmt.Sprintf("Could not find Datadog integration with ID '%s', name '%s'.", cfg.Id.ValueString(), cfg.Name.ValueString()),
	)
	return diags
}

func (d *datadogIntegrationDataSource) getByID(ctx context.Context, cfg datadogIntegrationDataSourceModel) (*schemas.DatadogIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	di, err := d.ClientV2.DatadogIntegration.GetDatadogIntegration(ctx, cfg.Id.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return nil, d.notFound(cfg)
		}
		diags.AddError("Error retrieving Datadog integration", err.Error())
		return nil, diags
	}

	if !cfg.Name.IsNull() && di.Attributes.Name != cfg.Name.ValueString() {
		return nil, d.notFound(cfg)
	}

	return di, diags
}

// findByName looks up the integration by name.
func (d *datadogIntegrationDataSource) findByName(ctx context.Context, cfg datadogIntegrationDataSourceModel) (*schemas.DatadogIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	opts := ddops.ListDatadogIntegrationsOptions{Filter: map[string]string{
		"name": cfg.Name.ValueString(),
	}}

	found, err := d.ClientV2.DatadogIntegration.ListDatadogIntegrations(ctx, &opts)
	if err != nil {
		diags.AddError("Error retrieving Datadog integration", err.Error())
		return nil, diags
	}

	switch len(found) {
	case 0:
		return nil, d.notFound(cfg)
	case 1:
		return found[0], diags
	default:
		diags.AddError(
			"Error retrieving Datadog integration",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return nil, diags
	}
}
