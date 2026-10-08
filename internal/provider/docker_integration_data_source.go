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
	"github.com/scalr/go-scalr/v2/scalr/ops/docker_integration"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &dockerIntegrationDataSource{}
	_ datasource.DataSourceWithConfigure        = &dockerIntegrationDataSource{}
	_ datasource.DataSourceWithConfigValidators = &dockerIntegrationDataSource{}
)

func newDockerIntegrationDataSource() datasource.DataSource {
	return &dockerIntegrationDataSource{}
}

// dockerIntegrationDataSource defines the data source implementation.
type dockerIntegrationDataSource struct {
	framework.DataSourceWithScalrClient
}

// dockerIntegrationDataSourceModel describes the data source data model.
type dockerIntegrationDataSourceModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	RegistryURL       types.String `tfsdk:"registry_url"`
	Username          types.String `tfsdk:"username"`
	ExportCredentials types.Bool   `tfsdk:"export_credentials"`
	Status            types.String `tfsdk:"status"`
}

func (d *dockerIntegrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_integration"
}

func (d *dockerIntegrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Docker (OCI) registry integration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Docker integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Docker integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"registry_url": schema.StringAttribute{
				MarkdownDescription: "The registry URL.",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The registry username.",
				Computed:            true,
			},
			"export_credentials": schema.BoolAttribute{
				MarkdownDescription: "Whether the registry credentials are exported into runs.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration after the latest connection test.",
				Computed:            true,
			},
		},
	}
}

func (d *dockerIntegrationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *dockerIntegrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg dockerIntegrationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var di *schemas.DockerIntegration
	var diags diag.Diagnostics

	if !cfg.Id.IsNull() {
		di, diags = d.findByID(ctx, cfg.Id.ValueString(), cfg.Name)
	} else {
		di, diags = d.findByName(ctx, cfg.Name.ValueString())
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg.Id = types.StringValue(di.ID)
	cfg.Name = types.StringValue(di.Attributes.Name)
	cfg.RegistryURL = types.StringValue(di.Attributes.RegistryUrl)
	cfg.Username = types.StringPointerValue(di.Attributes.Username)
	cfg.ExportCredentials = types.BoolValue(di.Attributes.ExportCredentials)
	cfg.Status = types.StringValue(string(di.Attributes.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

// findByID fetches the integration directly and, when a name is also
// configured, checks that it matches.
func (d *dockerIntegrationDataSource) findByID(
	ctx context.Context, id string, name types.String,
) (*schemas.DockerIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	di, err := d.ClientV2.DockerIntegration.GetDockerIntegration(ctx, id, nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			diags.AddError(
				"Error retrieving Docker integration",
				fmt.Sprintf("Could not find Docker integration with ID '%s'.", id),
			)
			return nil, diags
		}
		diags.AddError("Error retrieving Docker integration", err.Error())
		return nil, diags
	}

	if !name.IsNull() && di.Attributes.Name != name.ValueString() {
		diags.AddError(
			"Error retrieving Docker integration",
			fmt.Sprintf(
				"Could not find Docker integration with ID '%s', name '%s'.",
				id, name.ValueString(),
			),
		)
		return nil, diags
	}

	return di, diags
}

// findByName lists the integrations filtered by name.
func (d *dockerIntegrationDataSource) findByName(ctx context.Context, name string) (*schemas.DockerIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	opts := docker_integration.ListDockerIntegrationsOptions{Filter: map[string]string{"name": name}}
	matches, err := d.ClientV2.DockerIntegration.ListDockerIntegrations(ctx, &opts)
	if err != nil {
		diags.AddError("Error retrieving Docker integration", err.Error())
		return nil, diags
	}

	if len(matches) > 1 {
		diags.AddError(
			"Error retrieving Docker integration",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return nil, diags
	}

	if len(matches) == 0 {
		diags.AddError(
			"Error retrieving Docker integration",
			fmt.Sprintf("Could not find Docker integration with name '%s'.", name),
		)
		return nil, diags
	}

	return matches[0], diags
}
