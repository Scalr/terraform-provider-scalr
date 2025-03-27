package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

var (
	_ datasource.DataSource                     = &workloadIdentityProviderDataSource{}
	_ datasource.DataSourceWithConfigure        = &workloadIdentityProviderDataSource{}
	_ datasource.DataSourceWithConfigValidators = &workloadIdentityProviderDataSource{}
)

func newWorkloadIdentityProviderDataSource() datasource.DataSource {
	return &workloadIdentityProviderDataSource{}
}

type workloadIdentityProviderDataSource struct {
	framework.DataSourceWithScalrClient
}

type workloadIdentityProviderDataSourceModel struct {
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	URL                          types.String `tfsdk:"url"`
	AllowedAudiences             types.List   `tfsdk:"allowed_audiences"`
	CreatedAt                    types.String `tfsdk:"created_at"`
	CreatedByEmail               types.String `tfsdk:"created_by_email"`
	Status                       types.String `tfsdk:"status"`
	AssumeServiceAccountPolicies types.List   `tfsdk:"assume_service_account_policies"`
}

func (r *workloadIdentityProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workload_identity_provider"
}

func (r *workloadIdentityProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a single workload identity provider.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The workload identity provider ID.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the workload identity provider.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL of the workload identity provider.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"allowed_audiences": schema.ListAttribute{
				MarkdownDescription: "The list of allowed audiences for the workload identity provider.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the workload identity provider was created.",
				Computed:            true,
			},
			"created_by_email": schema.StringAttribute{
				MarkdownDescription: "The email of the user who created the workload identity provider.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the workload identity provider.",
				Computed:            true,
			},
			"assume_service_account_policies": schema.ListAttribute{
				MarkdownDescription: "The list of assume service account policy IDs associated with the workload identity provider.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *workloadIdentityProviderDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{}
}

func (r *workloadIdentityProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg workloadIdentityProviderDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := scalr.WorkloadIdentityProviderFilter{
		WorkloadIdentityProvider: cfg.ID.ValueString(),
		Name:                     cfg.Name.ValueString(),
		Url:                      cfg.URL.ValueString(),
	}

	options := scalr.WorkloadIdentityProvidersListOptions{
		Filter: &filter,
	}

	providers, err := r.Client.WorkloadIdentityProviders.List(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving workload identity providers", err.Error())
		return
	}

	if len(providers.Items) > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving workload identity provider",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if len(providers.Items) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving workload identity provider",
			fmt.Sprintf("Could not find workload identity provider with ID '%s', name '%s' or URL '%s'.",
				cfg.ID.ValueString(), cfg.Name.ValueString(), cfg.URL.ValueString()),
		)
		return
	}
	provider := providers.Items[0]

	cfg.ID = types.StringValue(provider.ID)
	cfg.Name = types.StringValue(provider.Name)
	cfg.URL = types.StringValue(provider.URL)
	cfg.CreatedAt = types.StringValue(provider.CreatedAt)
	cfg.Status = types.StringValue(provider.Status)

	if provider.CreatedByEmail != nil {
		cfg.CreatedByEmail = types.StringValue(*provider.CreatedByEmail)
	} else {
		cfg.CreatedByEmail = types.StringNull()
	}

	audiences, diags := types.ListValueFrom(context.Background(), types.StringType, provider.AllowedAudiences)
	if diags.HasError() {
		return
	}
	cfg.AllowedAudiences = audiences

	policies := make([]string, len(provider.AssumeServiceAccountPolicies))
	for i, policy := range provider.AssumeServiceAccountPolicies {
		policies[i] = policy.ID
	}
	policiesValue, diags := types.ListValueFrom(context.Background(), types.StringType, policies)
	if diags.HasError() {
		return
	}
	cfg.AssumeServiceAccountPolicies = policiesValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
