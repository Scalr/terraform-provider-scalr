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
	"github.com/scalr/go-scalr/v2/scalr/ops/checkov_integration"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &checkovIntegrationDataSource{}
	_ datasource.DataSourceWithConfigure        = &checkovIntegrationDataSource{}
	_ datasource.DataSourceWithConfigValidators = &checkovIntegrationDataSource{}
)

func newCheckovIntegrationDataSource() datasource.DataSource {
	return &checkovIntegrationDataSource{}
}

// checkovIntegrationDataSource defines the data source implementation.
type checkovIntegrationDataSource struct {
	framework.DataSourceWithScalrClient
}

// checkovIntegrationDataSourceModel describes the data source data model.
type checkovIntegrationDataSourceModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Version               types.String `tfsdk:"version"`
	CliArgs               types.String `tfsdk:"cli_args"`
	ScanMode              types.String `tfsdk:"scan_mode"`
	ExternalChecksEnabled types.Bool   `tfsdk:"external_checks_enabled"`
	Status                types.String `tfsdk:"status"`
	VCSProviderID         types.String `tfsdk:"vcs_provider_id"`
	VCSRepo               types.List   `tfsdk:"vcs_repo"`
	Environments          types.Set    `tfsdk:"environments"`
}

func (d *checkovIntegrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_checkov_integration"
}

func (d *checkovIntegrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Checkov integration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Checkov integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Checkov integration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the Checkov integration to use.",
				Computed:            true,
			},
			"cli_args": schema.StringAttribute{
				MarkdownDescription: "CLI parameters to be passed to checkov command.",
				Computed:            true,
			},
			"scan_mode": schema.StringAttribute{
				MarkdownDescription: "What Checkov scans: `configuration-files` for the configuration files, `plan-json` for the plan JSON.",
				Computed:            true,
			},
			"external_checks_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether external checks are enabled.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration.",
				Computed:            true,
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VCS provider the external checks are stored in, in the format `vcs-<RANDOM STRING>`.",
				Computed:            true,
			},
			"vcs_repo": schema.ListAttribute{
				MarkdownDescription: "Settings for the Checkov integration's VCS repository.",
				ElementType:         checkovVcsRepoElementType,
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

func (d *checkovIntegrationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *checkovIntegrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg checkovIntegrationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ci *schemas.CheckovIntegration
	var diags diag.Diagnostics

	if !cfg.Id.IsNull() {
		ci, diags = d.findByID(ctx, cfg.Id.ValueString(), cfg.Name)
	} else {
		ci, diags = d.findByName(ctx, cfg.Name.ValueString())
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg.Id = types.StringValue(ci.ID)
	cfg.Name = types.StringValue(ci.Attributes.Name)
	cfg.Version = types.StringValue(ci.Attributes.Version)
	cfg.CliArgs = types.StringValue(stringOrEmpty(ci.Attributes.CliArgs))
	cfg.ScanMode = types.StringValue(string(ci.Attributes.ScanMode))
	cfg.ExternalChecksEnabled = types.BoolValue(ci.Attributes.ExternalChecksEnabled)
	cfg.Status = types.StringValue(string(ci.Attributes.Status))

	cfg.VCSProviderID = types.StringNull()
	if ci.Relationships.VcsProvider != nil {
		cfg.VCSProviderID = types.StringValue(ci.Relationships.VcsProvider.ID)
	}

	cfg.VCSRepo = types.ListNull(checkovVcsRepoElementType)
	if ci.Attributes.VcsRepo != nil {
		repo := []checkovVcsRepoModel{{
			Identifier: types.StringValue(ci.Attributes.VcsRepo.Identifier),
			Branch:     types.StringValue(ci.Attributes.VcsRepo.Branch),
			Path:       types.StringValue(stringOrEmpty(ci.Attributes.VcsRepo.Path)),
		}}
		repoValue, d := types.ListValueFrom(ctx, checkovVcsRepoElementType, repo)
		resp.Diagnostics.Append(d...)
		cfg.VCSRepo = repoValue
	}

	var envs types.Set
	if ci.Attributes.IsShared {
		envs, diags = types.SetValueFrom(ctx, types.StringType, []string{"*"})
	} else {
		envs, diags = framework.FlattenRelationshipIDsSet(
			ctx,
			ci.Relationships.Environments,
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
func (d *checkovIntegrationDataSource) findByID(
	ctx context.Context, id string, name types.String,
) (*schemas.CheckovIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	ci, err := d.ClientV2.CheckovIntegration.GetCheckovIntegration(ctx, id, nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			diags.AddError(
				"Error retrieving Checkov integration",
				fmt.Sprintf("Could not find Checkov integration with ID '%s'.", id),
			)
			return nil, diags
		}
		diags.AddError("Error retrieving Checkov integration", err.Error())
		return nil, diags
	}

	if !name.IsNull() && ci.Attributes.Name != name.ValueString() {
		diags.AddError(
			"Error retrieving Checkov integration",
			fmt.Sprintf(
				"Could not find Checkov integration with ID '%s', name '%s'.",
				id, name.ValueString(),
			),
		)
		return nil, diags
	}

	return ci, diags
}

// findByName lists the integrations filtered by name.
func (d *checkovIntegrationDataSource) findByName(ctx context.Context, name string) (*schemas.CheckovIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics

	opts := checkov_integration.ListCheckovIntegrationsOptions{Filter: map[string]string{"name": name}}
	matches, err := d.ClientV2.CheckovIntegration.ListCheckovIntegrations(ctx, &opts)
	if err != nil {
		diags.AddError("Error retrieving Checkov integration", err.Error())
		return nil, diags
	}

	if len(matches) > 1 {
		diags.AddError(
			"Error retrieving Checkov integration",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return nil, diags
	}

	if len(matches) == 0 {
		diags.AddError(
			"Error retrieving Checkov integration",
			fmt.Sprintf("Could not find Checkov integration with name '%s'.", name),
		)
		return nil, diags
	}

	return matches[0], diags
}
