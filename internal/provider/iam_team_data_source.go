package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/ops/team"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &iamTeamDataSource{}
	_ datasource.DataSourceWithConfigure        = &iamTeamDataSource{}
	_ datasource.DataSourceWithConfigValidators = &iamTeamDataSource{}
)

func newIamTeamDataSource() datasource.DataSource {
	return &iamTeamDataSource{}
}

// iamTeamDataSource defines the data source implementation.
type iamTeamDataSource struct {
	framework.DataSourceWithScalrClient
}

// iamTeamDataSourceModel describes the data source data model.
type iamTeamDataSourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	AccountID          types.String `tfsdk:"account_id"`
	Description        types.String `tfsdk:"description"`
	IdentityProviderID types.String `tfsdk:"identity_provider_id"`
	Users              types.Set    `tfsdk:"users"`
}

func (d *iamTeamDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_iam_team"
}

func (d *iamTeamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the details of a Scalr team.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the team.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the team.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A verbose description of the team.",
				Computed:            true,
			},
			"identity_provider_id": schema.StringAttribute{
				MarkdownDescription: "An identifier of an identity provider team is linked to, in the format `idp-<RANDOM STRING>`.",
				Computed:            true,
			},
			"users": schema.SetAttribute{
				MarkdownDescription: "The list of the user identifiers that belong to the team.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *iamTeamDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *iamTeamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg iamTeamDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := team.GetTeamsOptions{Filter: map[string]string{}}
	if !cfg.Id.IsUnknown() && !cfg.Id.IsNull() {
		opts.Filter["team"] = cfg.Id.ValueString()
	}
	if !cfg.Name.IsUnknown() && !cfg.Name.IsNull() {
		opts.Filter["name"] = cfg.Name.ValueString()
	}

	iamTeams, err := d.ClientV2.Team.GetTeams(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving team", err.Error())
		return
	}

	if len(iamTeams) > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving team",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if len(iamTeams) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving team",
			"Could not find a team matching the provided criteria.",
		)
		return
	}

	t := iamTeams[0]

	cfg.Id = types.StringValue(t.ID)
	cfg.Name = types.StringValue(t.Attributes.Name)
	cfg.Description = types.StringNull()
	cfg.AccountID = types.StringNull()
	cfg.IdentityProviderID = types.StringNull()
	cfg.Users = types.SetNull(types.StringType)

	if t.Attributes.Description != nil {
		cfg.Description = types.StringValue(*t.Attributes.Description)
	}
	if t.Relationships.Account != nil {
		cfg.AccountID = types.StringValue(t.Relationships.Account.ID)
	}
	if t.Relationships.IdentityProvider != nil {
		cfg.IdentityProviderID = types.StringValue(t.Relationships.IdentityProvider.ID)
	}
	users, diags := framework.FlattenRelationshipIDsSet(
		ctx,
		t.Relationships.Users,
		func(u *schemas.User) string { return u.ID },
		nil,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg.Users = users

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
