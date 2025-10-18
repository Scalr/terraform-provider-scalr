package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &providerConfigurationDataSource{}
	_ datasource.DataSourceWithConfigure        = &providerConfigurationDataSource{}
	_ datasource.DataSourceWithConfigValidators = &providerConfigurationDataSource{}
)

func newProviderConfigurationDataSource() datasource.DataSource {
	return &providerConfigurationDataSource{}
}

// providerConfigurationDataSource defines the data source implementation.
type providerConfigurationDataSource struct {
	framework.DataSourceWithScalrClient
}

// providerConfigurationDataSourceModel describes the data source data model.
type providerConfigurationDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProviderName types.String `tfsdk:"provider_name"`
	Environments types.List   `tfsdk:"environments"`
	Owners       types.List   `tfsdk:"owners"`
	AccountID    types.String `tfsdk:"account_id"`
}

func (r *providerConfigurationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_configuration"
}

func (r *providerConfigurationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a single provider configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The provider configuration ID, in the format `pcfg-xxxxxxxxxxx`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of a Scalr provider configuration.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"provider_name": schema.StringAttribute{
				MarkdownDescription: "The name of a Terraform provider.",
				Optional:            true,
			},
			"environments": schema.ListAttribute{
				MarkdownDescription: "The list of environment identifiers that the provider configuration is shared to, or `[\"*\"]` if shared with all environments.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"owners": schema.ListAttribute{
				MarkdownDescription: "The teams, the provider configuration belongs to.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *providerConfigurationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{}
}

func (r *providerConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg providerConfigurationDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	providersFilter := scalr.ProviderConfigurationFilter{
		ProviderConfiguration: cfg.Id.ValueString(),
		AccountID:             cfg.AccountID.ValueString(),
		Name:                  cfg.Name.ValueString(),
		ProviderName:          cfg.ProviderName.ValueString(),
	}
	options := scalr.ProviderConfigurationsListOptions{
		Filter: &providersFilter,
	}

	providerConfigurations, err := r.Client.ProviderConfigurations.List(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider_configuration", err.Error())
		return
	}

	if len(providerConfigurations.Items) > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving provider_configuration",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if len(providerConfigurations.Items) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving provider_configuration",
			fmt.Sprintf("Could not find provider_configuration with ID '%s', name '%s' and provider_name '%s'.", cfg.Id.ValueString(), cfg.Name.ValueString(), cfg.ProviderName.ValueString()),
		)
		return
	}

	providerConfiguration := providerConfigurations.Items[0]

	cfg.Id = types.StringValue(providerConfiguration.ID)
	cfg.Name = types.StringValue(providerConfiguration.Name)
	cfg.ProviderName = types.StringValue(providerConfiguration.ProviderName)
	cfg.AccountID = types.StringValue(providerConfiguration.Account.ID)

	owners := make([]string, len(providerConfiguration.Owners))
	for i, owner := range providerConfiguration.Owners {
		owners[i] = owner.ID
	}
	sort.Strings(owners)
	ownersValue, d := types.ListValueFrom(ctx, types.StringType, owners)
	resp.Diagnostics.Append(d...)
	cfg.Owners = ownersValue

	var environments []string
	if providerConfiguration.IsShared {
		environments = []string{"*"}
	} else {
		environments = make([]string, len(providerConfiguration.Environments))
		for i, environment := range providerConfiguration.Environments {
			environments[i] = environment.ID
		}
		sort.Strings(environments)
	}
	environmentsValue, d := types.ListValueFrom(ctx, types.StringType, environments)
	resp.Diagnostics.Append(d...)
	cfg.Environments = environmentsValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
