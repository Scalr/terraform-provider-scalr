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
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &wizIntegrationDataSource{}
	_ datasource.DataSourceWithConfigure        = &wizIntegrationDataSource{}
	_ datasource.DataSourceWithConfigValidators = &wizIntegrationDataSource{}
)

func newWizIntegrationDataSource() datasource.DataSource {
	return &wizIntegrationDataSource{}
}

// wizIntegrationDataSource defines the data source implementation.
type wizIntegrationDataSource struct {
	framework.DataSourceWithScalrClient
}

// wizIntegrationDataSourceModel describes the data source data model.
type wizIntegrationDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ClientID        types.String `tfsdk:"client_id"`
	DefaultPolicies types.Set    `tfsdk:"default_policies"`
	DefaultScanName types.String `tfsdk:"default_scan_name"`
	Autofail        types.Bool   `tfsdk:"autofail"`
	EndpointMode    types.String `tfsdk:"endpoint_mode"`
	Status          types.String `tfsdk:"status"`
	Environments    types.Set    `tfsdk:"environments"`
}

func (d *wizIntegrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wiz_integration"
}

func (d *wizIntegrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Wiz integration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Wiz integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Wiz integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Wiz service-account client ID.",
				Computed:            true,
			},
			"default_policies": schema.SetAttribute{
				MarkdownDescription: "Wiz CI/CD policy names passed to `wizcli` (`--policies`), one list item per policy.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"default_scan_name": schema.StringAttribute{
				MarkdownDescription: "Default scan name passed to `wizcli` (`--name`).",
				Computed:            true,
			},
			"autofail": schema.BoolAttribute{
				MarkdownDescription: "Whether the run is blocked when the Wiz scan fails or cannot run.",
				Computed:            true,
			},
			"endpoint_mode": schema.StringAttribute{
				MarkdownDescription: "Wiz tenant region: `commercial`, `govcloud`, or `fedramp`.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration.",
				Computed:            true,
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of environments this integration is linked to, or `[\"*\"]` if shared with all environments.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *wizIntegrationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *wizIntegrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg wizIntegrationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var wi *schemas.WizIntegration
	var diags diag.Diagnostics

	if !cfg.Id.IsNull() {
		wi, diags = d.findByID(ctx, cfg.Id.ValueString(), cfg.Name)
	} else {
		wi, diags = d.findByName(ctx, cfg.Name.ValueString())
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg.Id = types.StringValue(wi.ID)
	cfg.Name = types.StringValue(wi.Attributes.Name)
	cfg.ClientID = types.StringValue(wi.Attributes.ClientId)
	cfg.DefaultScanName = types.StringPointerValue(wi.Attributes.DefaultScanName)
	cfg.Autofail = types.BoolValue(wi.Attributes.Autofail)
	cfg.EndpointMode = types.StringValue(string(wi.Attributes.EndpointMode))
	cfg.Status = types.StringValue(string(wi.Attributes.Status))

	policies := []string{}
	if wi.Attributes.DefaultPolicies != nil {
		policies = *wi.Attributes.DefaultPolicies
	}
	policiesSet, diags := types.SetValueFrom(ctx, types.StringType, policies)
	resp.Diagnostics.Append(diags...)
	cfg.DefaultPolicies = policiesSet

	var envs types.Set
	if wi.Attributes.IsShared {
		envs, diags = types.SetValueFrom(ctx, types.StringType, []string{"*"})
	} else {
		envs, diags = framework.FlattenRelationshipIDsSet(
			ctx,
			wi.Relationships.Environments,
			func(e *schemas.Environment) string { return e.ID },
			nil,
		)
	}
	resp.Diagnostics.Append(diags...)
	cfg.Environments = envs

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

// findByID fetches the integration directly and, when a name is also
// configured, checks that it matches.
func (d *wizIntegrationDataSource) findByID(
	ctx context.Context, id string, name types.String,
) (*schemas.WizIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	wi, err := d.ClientV2.WizIntegration.GetWizIntegration(ctx, id, nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			diags.AddError(
				"Error retrieving Wiz integration",
				fmt.Sprintf("Could not find Wiz integration with ID '%s'.", id),
			)
			return nil, diags
		}
		diags.AddError("Error retrieving Wiz integration", err.Error())
		return nil, diags
	}

	if !name.IsNull() && wi.Attributes.Name != name.ValueString() {
		diags.AddError(
			"Error retrieving Wiz integration",
			fmt.Sprintf(
				"Could not find Wiz integration with ID '%s', name '%s'.",
				id, name.ValueString(),
			),
		)
		return nil, diags
	}

	return wi, diags
}

// findByName lists the integrations and matches on name, since the API does
// not support filtering this collection. The backend allows at most one Wiz
// integration per account, so a single page is always enough.
func (d *wizIntegrationDataSource) findByName(ctx context.Context, name string) (*schemas.WizIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	integrations, err := d.ClientV2.WizIntegration.ListWizIntegrations(ctx, nil)
	if err != nil {
		diags.AddError("Error retrieving Wiz integration", err.Error())
		return nil, diags
	}

	var matches []*schemas.WizIntegration
	for _, wi := range integrations {
		if wi.Attributes.Name == name {
			matches = append(matches, wi)
		}
	}

	if len(matches) > 1 {
		diags.AddError(
			"Error retrieving Wiz integration",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return nil, diags
	}

	if len(matches) == 0 {
		diags.AddError(
			"Error retrieving Wiz integration",
			fmt.Sprintf("Could not find Wiz integration with name '%s'.", name),
		)
		return nil, diags
	}

	return matches[0], diags
}
