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
	splunkops "github.com/scalr/go-scalr/v2/scalr/ops/splunk_integration"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &splunkIntegrationDataSource{}
	_ datasource.DataSourceWithConfigure        = &splunkIntegrationDataSource{}
	_ datasource.DataSourceWithConfigValidators = &splunkIntegrationDataSource{}
)

func newSplunkIntegrationDataSource() datasource.DataSource {
	return &splunkIntegrationDataSource{}
}

// splunkIntegrationDataSource defines the data source implementation.
type splunkIntegrationDataSource struct {
	framework.DataSourceWithScalrClient
}

// splunkIntegrationDataSourceModel describes the data source data model.
type splunkIntegrationDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Endpoint   types.String `tfsdk:"endpoint"`
	Index      types.String `tfsdk:"index"`
	Status     types.String `tfsdk:"status"`
	ErrMessage types.String `tfsdk:"err_message"`
	AccountID  types.String `tfsdk:"account_id"`
}

func (d *splunkIntegrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_splunk_integration"
}

func (d *splunkIntegrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Splunk integration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Splunk integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Splunk integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base HTTP Event Collector URL.",
				Computed:            true,
			},
			"index": schema.StringAttribute{
				MarkdownDescription: "Splunk index for the events. If not set, events go to the default index of the HEC token.",
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

func (d *splunkIntegrationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *splunkIntegrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg splunkIntegrationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var si *schemas.SplunkIntegration
	var diags diag.Diagnostics
	if !cfg.Id.IsNull() {
		si, diags = d.getByID(ctx, cfg)
	} else {
		si, diags = d.findByName(ctx, cfg)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg.Id = types.StringValue(si.ID)
	cfg.Name = types.StringValue(si.Attributes.Name)
	cfg.Endpoint = types.StringValue(si.Attributes.Endpoint)
	cfg.Index = types.StringPointerValue(si.Attributes.Index)
	cfg.Status = types.StringValue(string(si.Attributes.Status))
	cfg.ErrMessage = types.StringPointerValue(si.Attributes.ErrMessage)
	if si.Relationships.Account != nil {
		cfg.AccountID = types.StringValue(si.Relationships.Account.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

func (d *splunkIntegrationDataSource) notFound(cfg splunkIntegrationDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Error retrieving Splunk integration",
		fmt.Sprintf("Could not find Splunk integration with ID '%s', name '%s'.", cfg.Id.ValueString(), cfg.Name.ValueString()),
	)
	return diags
}

func (d *splunkIntegrationDataSource) getByID(ctx context.Context, cfg splunkIntegrationDataSourceModel) (*schemas.SplunkIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	si, err := d.ClientV2.SplunkIntegration.GetSplunkIntegration(ctx, cfg.Id.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return nil, d.notFound(cfg)
		}
		diags.AddError("Error retrieving Splunk integration", err.Error())
		return nil, diags
	}

	if !cfg.Name.IsNull() && si.Attributes.Name != cfg.Name.ValueString() {
		return nil, d.notFound(cfg)
	}

	return si, diags
}

func (d *splunkIntegrationDataSource) findByName(ctx context.Context, cfg splunkIntegrationDataSourceModel) (*schemas.SplunkIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	opts := splunkops.ListSplunkIntegrationsOptions{Filter: map[string]string{
		"name": cfg.Name.ValueString(),
	}}

	found, err := d.ClientV2.SplunkIntegration.ListSplunkIntegrations(ctx, &opts)
	if err != nil {
		diags.AddError("Error retrieving Splunk integration", err.Error())
		return nil, diags
	}

	switch len(found) {
	case 0:
		return nil, d.notFound(cfg)
	case 1:
		return found[0], diags
	default:
		diags.AddError(
			"Error retrieving Splunk integration",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return nil, diags
	}
}
