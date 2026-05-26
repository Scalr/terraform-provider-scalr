package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	varsetops "github.com/scalr/go-scalr/v2/scalr/ops/variable_set"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &varSetDataSource{}
	_ datasource.DataSourceWithConfigure        = &varSetDataSource{}
	_ datasource.DataSourceWithConfigValidators = &varSetDataSource{}
)

func newVarSetDataSource() datasource.DataSource {
	return &varSetDataSource{}
}

// varSetDataSource defines the data source implementation.
type varSetDataSource struct {
	framework.DataSourceWithScalrClient
}

// varSetDataSourceModel describes the data source data model.
type varSetDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Environments   types.Set    `tfsdk:"environments"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	UpdatedByEmail types.String `tfsdk:"updated_by_email"`
	AccountID      types.String `tfsdk:"account_id"`
	Owners         types.Set    `tfsdk:"owners"`
}

func (d *varSetDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_var_set"
}

func (d *varSetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about variable set.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the variable set.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the variable set.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the variable set.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "UTC timestamp of the last update to this variable set.",
				Computed:            true,
			},
			"updated_by_email": schema.StringAttribute{
				MarkdownDescription: "Email of the user who last updated this variable set.",
				Computed:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account this variable set belongs to.",
				Computed:            true,
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of environment IDs that this variable set is shared to. `[\"*\"]` means shared with all environments.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"owners": schema.SetAttribute{
				MarkdownDescription: "List of team IDs this variable set belongs to.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *varSetDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *varSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg varSetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := varsetops.ListVarSetsOptions{Filter: map[string]string{}}
	if !cfg.Id.IsUnknown() && !cfg.Id.IsNull() {
		opts.Filter["var-set"] = cfg.Id.ValueString()
	}
	if !cfg.Name.IsUnknown() && !cfg.Name.IsNull() {
		opts.Filter["name"] = cfg.Name.ValueString()
	}

	varSets, err := d.ClientV2.VariableSet.ListVarSets(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving var_set", err.Error())
		return
	}

	if len(varSets) > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving var_set",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if len(varSets) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving var_set",
			"Could not find variable set matching the provided criteria.",
		)
		return
	}

	vs := varSets[0]

	cfg.Id = types.StringValue(vs.ID)
	cfg.Name = types.StringValue(vs.Attributes.Name)
	cfg.Description = types.StringPointerValue(vs.Attributes.Description)
	cfg.UpdatedAt = types.StringValue(vs.Attributes.UpdatedAt.Format(time.RFC3339))
	cfg.UpdatedByEmail = types.StringPointerValue(vs.Attributes.UpdatedByEmail)

	if vs.Relationships.Account != nil {
		cfg.AccountID = types.StringValue(vs.Relationships.Account.ID)
	}

	owners, diags := framework.FlattenRelationshipIDsSet(
		ctx,
		vs.Relationships.Owners,
		func(t *schemas.Team) string { return t.ID },
		nil,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg.Owners = owners

	var envs types.Set
	if vs.Attributes.IsShared {
		envs, diags = types.SetValueFrom(ctx, types.StringType, []string{"*"})
	} else {
		envs, diags = framework.FlattenRelationshipIDsSet(
			ctx,
			vs.Relationships.Environments,
			func(e *schemas.Environment) string { return e.ID },
			nil,
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg.Environments = envs

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
