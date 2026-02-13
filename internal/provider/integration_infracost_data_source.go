package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &integrationInfracostDataSource{}
	_ datasource.DataSourceWithConfigure        = &integrationInfracostDataSource{}
	_ datasource.DataSourceWithConfigValidators = &integrationInfracostDataSource{}
)

func newIntegrationInfracostDataSource() datasource.DataSource {
	return &integrationInfracostDataSource{}
}

// integrationInfracostDataSource defines the data source implementation.
type integrationInfracostDataSource struct {
	framework.DataSourceWithScalrClient
}

// integrationInfracostDataSourceModel describes the data source data model.
type integrationInfracostDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Environments types.Set    `tfsdk:"environments"`
}

func (d *integrationInfracostDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_infracost"
}

func (d *integrationInfracostDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about Infracost integration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Infracost integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Infracost integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of environments this integration is linked to, or `[\"*\"]` if shared with all environments.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *integrationInfracostDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *integrationInfracostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg integrationInfracostDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.InfracostIntegrationListOptions{}
	if !cfg.Id.IsNull() {
		opts.InfracostIntegration = cfg.Id.ValueStringPointer()
	}
	if !cfg.Name.IsNull() {
		opts.Name = cfg.Name.ValueStringPointer()
	}

	integrationInfracosts, err := d.Client.InfracostIntegrations.List(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Infracost integration", err.Error())
		return
	}

	// Unlikely
	if integrationInfracosts.TotalCount > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving Infracost integration",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if integrationInfracosts.TotalCount == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving Infracost integration",
			fmt.Sprintf("Could not find Infracost integration with ID '%s', name '%s'.", cfg.Id.ValueString(), cfg.Name.ValueString()),
		)
		return
	}

	integrationInfracost := integrationInfracosts.Items[0]

	cfg.Id = types.StringValue(integrationInfracost.ID)
	cfg.Name = types.StringValue(integrationInfracost.Name)
	if integrationInfracost.IsShared {
		envs := []string{"*"}
		envsValues, _ := types.SetValueFrom(ctx, types.StringType, envs)
		cfg.Environments = envsValues
	} else {
		envs := make([]string, len(integrationInfracost.Environments))
		for i, env := range integrationInfracost.Environments {
			envs[i] = env.ID
		}
		envsValues, _ := types.SetValueFrom(ctx, types.StringType, envs)
		cfg.Environments = envsValues
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
