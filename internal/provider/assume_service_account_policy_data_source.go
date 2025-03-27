package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

var (
	_ datasource.DataSource                     = &assumeServiceAccountPolicyDataSource{}
	_ datasource.DataSourceWithConfigure        = &assumeServiceAccountPolicyDataSource{}
	_ datasource.DataSourceWithConfigValidators = &assumeServiceAccountPolicyDataSource{}
)

func newAssumeServiceAccountPolicyDataSource() datasource.DataSource {
	return &assumeServiceAccountPolicyDataSource{}
}

type assumeServiceAccountPolicyDataSource struct {
	framework.DataSourceWithScalrClient
}

type assumeServiceAccountPolicyDataSourceModel struct {
	ID                     types.String          `tfsdk:"id"`
	Name                   types.String          `tfsdk:"name"`
	ServiceAccountID       types.String          `tfsdk:"service_account_id"`
	ProviderID             types.String          `tfsdk:"provider_id"`
	MaximumSessionDuration types.Int64           `tfsdk:"maximum_session_duration"`
	ClaimConditions        []claimConditionModel `tfsdk:"claim_conditions"`
	CreatedAt              types.String          `tfsdk:"created_at"`
	CreatedByEmail         types.String          `tfsdk:"created_by_email"`
}

func (d *assumeServiceAccountPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assume_service_account_policy"
}

func (d *assumeServiceAccountPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a Scalr Assume Service Account Policy.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the policy.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the policy.",
				Optional:    true,
				Computed:    true,
			},
			"service_account_id": schema.StringAttribute{
				Description: "The ID of the service account this policy belongs to.",
				Required:    true,
			},
			"provider_id": schema.StringAttribute{
				Description: "The ID of the workload identity provider.",
				Optional:    true,
				Computed:    true,
			},
			"maximum_session_duration": schema.Int64Attribute{
				Description: "Maximum session duration in seconds.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the policy was created.",
				Computed:    true,
			},
			"created_by_email": schema.StringAttribute{
				Description: "Email of the user who created the policy.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"claim_conditions": schema.ListNestedBlock{
				Description: "Conditions that must be met for the policy to be assumed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"claim": schema.StringAttribute{
							Description: "The claim to match against.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The value to match against.",
							Computed:    true,
						},
						"operator": schema.StringAttribute{
							Description: "The operator to use for matching ('eq', 'like', 'startswith', or 'endswith').",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *assumeServiceAccountPolicyDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{}
}

func (d *assumeServiceAccountPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config assumeServiceAccountPolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := scalr.AssumeServiceAccountPoliciesListOptions{
		Filter: &scalr.AssumeServiceAccountPolicyFilter{
			AssumeServiceAccountPolicy: config.ID.ValueString(),
			Name:                       config.Name.ValueString(),
			ServiceAccount:             config.ServiceAccountID.ValueString(),
			WorkloadIdentityProvider:   config.ProviderID.ValueString(),
		},
	}

	policies, err := d.Client.AssumeServiceAccountPolicies.List(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing assume service account policies",
			err.Error(),
		)
		return
	}

	if len(policies.Items) > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving assume service account policy",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if len(policies.Items) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving assume service account policy",
			fmt.Sprintf("Could not find policy with name '%s' for service account '%s'",
				config.Name.ValueString(), config.ServiceAccountID.ValueString()),
		)
		return
	}
	policy := policies.Items[0]

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.ServiceAccountID = types.StringValue(policy.ServiceAccount.ID)
	config.ProviderID = types.StringValue(policy.Provider.ID)
	config.MaximumSessionDuration = types.Int64Value(int64(policy.MaximumSessionDuration))
	config.CreatedAt = types.StringValue(policy.CreatedAt)

	if policy.CreatedByEmail != nil {
		config.CreatedByEmail = types.StringValue(*policy.CreatedByEmail)
	} else {
		config.CreatedByEmail = types.StringNull()
	}

	config.ClaimConditions = make([]claimConditionModel, len(policy.ClaimConditions))
	for i, cc := range policy.ClaimConditions {
		config.ClaimConditions[i] = claimConditionModel{
			Claim:    types.StringValue(cc.Claim),
			Value:    types.StringValue(cc.Value),
			Operator: stringValueOrNull(cc.Operator),
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func stringValueOrNull(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
