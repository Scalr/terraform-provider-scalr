package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/ops/policy_group"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &policyGroupDataSource{}
	_ datasource.DataSourceWithConfigure        = &policyGroupDataSource{}
	_ datasource.DataSourceWithConfigValidators = &policyGroupDataSource{}
)

func newPolicyGroupDataSource() datasource.DataSource {
	return &policyGroupDataSource{}
}

// policyGroupDataSource defines the data source implementation.
type policyGroupDataSource struct {
	framework.DataSourceWithScalrClient
}

// policyGroupDataSourceModel describes the data source data model.
type policyGroupDataSourceModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Status                types.String `tfsdk:"status"`
	ErrorMessage          types.String `tfsdk:"error_message"`
	OpaVersion            types.String `tfsdk:"opa_version"`
	EvaluateOn            types.String `tfsdk:"evaluate_on"`
	VCSRepo               types.List   `tfsdk:"vcs_repo"`
	CommonFunctionsFolder types.String `tfsdk:"common_functions_folder"`
	AccountID             types.String `tfsdk:"account_id"`
	VCSProviderID         types.String `tfsdk:"vcs_provider_id"`
	Policies              types.List   `tfsdk:"policies"`
	Environments          types.List   `tfsdk:"environments"`
}

func (d *policyGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_group"
}

func (d *policyGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the details of a policy group.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a policy group.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of a policy group.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "A system status of the policy group.",
				Computed:            true,
			},
			"error_message": schema.StringAttribute{
				MarkdownDescription: "An error details if Scalr failed to process the policy group.",
				Computed:            true,
			},
			"opa_version": schema.StringAttribute{
				MarkdownDescription: "The version of the Open Policy Agent that the policy group is using.",
				Computed:            true,
			},
			"evaluate_on": schema.StringAttribute{
				MarkdownDescription: "The stage of the run the policy group is evaluated at: `pre-plan` or `post-plan`.",
				Computed:            true,
			},
			"vcs_repo": schema.ListAttribute{
				MarkdownDescription: "Contains VCS-related meta-data for the policy group.",
				ElementType:         policyGroupVcsRepoElementType,
				Computed:            true,
			},
			"common_functions_folder": schema.StringAttribute{
				MarkdownDescription: "An absolute path from the repository root to the folder that contains common rego functions.",
				Computed:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Scalr account.",
				Optional:            true,
				Computed:            true,
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "The VCS provider identifier for the repository where the policy group resides. In the format `vcs-<RANDOM STRING>`.",
				Computed:            true,
			},
			"policies": schema.ListAttribute{
				MarkdownDescription: "A list of the OPA policies the policy group verifies each run.",
				ElementType:         policyGroupPolicyElementType,
				Computed:            true,
			},
			"environments": schema.ListAttribute{
				MarkdownDescription: "A list of the environments the policy group is linked to, or `[\"*\"]` if enforced in all environments.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *policyGroupDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *policyGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg policyGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := cfg.AccountID.ValueString()
	if cfg.AccountID.IsNull() {
		var diags diag.Diagnostics
		accountID, diags = defaults.GetDefaultScalrAccountID()
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	opts := &policy_group.ListPolicyGroupsOptions{
		Include: []string{"policies"},
		Filter:  map[string]string{"account": accountID},
	}
	if !cfg.Id.IsNull() {
		opts.Filter["policy-group"] = cfg.Id.ValueString()
	}
	if !cfg.Name.IsNull() {
		opts.Filter["name"] = cfg.Name.ValueString()
	}

	pgs, err := d.ClientV2.PolicyGroup.ListPolicyGroups(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving policy group", err.Error())
		return
	}
	if len(pgs) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving policy group",
			fmt.Sprintf(
				"Policy group with ID '%s', name '%s' and account_id '%s' not found",
				cfg.Id.ValueString(), cfg.Name.ValueString(), accountID,
			),
		)
		return
	}

	pg, diags := policyGroupResourceModelFromAPI(ctx, pgs[0])
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var environments []string
	resp.Diagnostics.Append(pg.Environments.ElementsAs(ctx, &environments, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sort.Strings(environments)
	environmentsValue, diags := types.ListValueFrom(ctx, types.StringType, environments)
	resp.Diagnostics.Append(diags...)

	cfg.Id = pg.Id
	cfg.Name = pg.Name
	cfg.Status = pg.Status
	cfg.ErrorMessage = pg.ErrorMessage
	cfg.OpaVersion = pg.OpaVersion
	cfg.EvaluateOn = pg.EvaluateOn
	cfg.VCSRepo = pg.VCSRepo
	cfg.CommonFunctionsFolder = pg.CommonFunctionsFolder
	cfg.AccountID = types.StringValue(accountID)
	cfg.VCSProviderID = pg.VCSProviderID
	cfg.Policies = pg.Policies
	cfg.Environments = environmentsValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
