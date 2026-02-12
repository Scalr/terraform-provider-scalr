package provider

import (
	"context"
	"fmt"
	"sort"

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
	_ datasource.DataSource                     = &environmentDataSource{}
	_ datasource.DataSourceWithConfigure        = &environmentDataSource{}
	_ datasource.DataSourceWithConfigValidators = &environmentDataSource{}
)

func newEnvironmentDataSource() datasource.DataSource {
	return &environmentDataSource{}
}

// environmentDataSource defines the data source implementation.
type environmentDataSource struct {
	framework.DataSourceWithScalrClient
}

// environmentDataSourceModel describes the data source data model.
type environmentDataSourceModel struct {
	Id                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	Status                        types.String `tfsdk:"status"`
	CreatedBy                     types.List   `tfsdk:"created_by"`
	PolicyGroups                  types.List   `tfsdk:"policy_groups"`
	TagIDs                        types.List   `tfsdk:"tag_ids"`
	DefaultProviderConfigurations types.List   `tfsdk:"default_provider_configurations"`
	RemoteBackend                 types.Bool   `tfsdk:"remote_backend"`
	RemoteBackendOverridable      types.Bool   `tfsdk:"remote_backend_overridable"`
	MaskSensitiveOutput           types.Bool   `tfsdk:"mask_sensitive_output"`
	FederatedEnvironments         types.Set    `tfsdk:"federated_environments"`
	AccountID                     types.String `tfsdk:"account_id"`
	StorageProfileID              types.String `tfsdk:"storage_profile_id"`
	DefaultWorkspaceAgentPoolID   types.String `tfsdk:"default_workspace_agent_pool_id"`
}

func (d *environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about environment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The environment ID, in the format `env-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the environment.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of an environment.",
				Computed:            true,
			},
			"created_by": schema.ListAttribute{
				MarkdownDescription: "Details of the user that created the environment.",
				ElementType:         userElementType,
				Computed:            true,
			},
			"policy_groups": schema.ListAttribute{
				MarkdownDescription: "List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"tag_ids": schema.ListAttribute{
				MarkdownDescription: "List of tag IDs associated with the environment.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"default_provider_configurations": schema.ListAttribute{
				MarkdownDescription: "List of IDs of provider configurations, used in the environment workspaces by default.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"remote_backend": schema.BoolAttribute{
				MarkdownDescription: "If Scalr exports the remote backend configuration and state storage for your infrastructure management." +
					" Disabling this feature will also prevent the ability to perform state locking, which ensures that concurrent operations do not conflict." +
					" Additionally, it will disable the capability to initiate CLI-driven runs through Scalr.",
				Computed: true,
			},
			"remote_backend_overridable": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the remote backend configuration can be overridden on the workspace level.",
				Computed:            true,
			},
			"mask_sensitive_output": schema.BoolAttribute{
				MarkdownDescription: "Enable masking of the sensitive console output.",
				Computed:            true,
			},
			"federated_environments": schema.SetAttribute{
				MarkdownDescription: "The list of environment identifiers that are allowed to access this environment, or `[\"*\"]` if shared with all environments.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
			},
			"storage_profile_id": schema.StringAttribute{
				MarkdownDescription: "The storage profile for this environment.",
				Computed:            true,
			},
			"default_workspace_agent_pool_id": schema.StringAttribute{
				MarkdownDescription: "Default agent pool that will be set for the entire environment. It will be used by a workspace if no other pool is explicitly linked.",
				Computed:            true,
			},
		},
	}
}

func (d *environmentDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg environmentDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var environment *scalr.Environment
	var err error

	if !cfg.Id.IsNull() {
		environment, err = d.Client.Environments.Read(ctx, cfg.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving environment", err.Error())
			return
		}

		if !cfg.Name.IsNull() && cfg.Name.ValueString() != environment.Name {
			resp.Diagnostics.AddError(
				"Error retrieving environment",
				fmt.Sprintf("Could not find environment with ID '%s' and name '%s'", cfg.Id.ValueString(), cfg.Name.ValueString()),
			)
			return
		}
	} else {
		options := GetEnvironmentByNameOptions{
			Name:    cfg.Name.ValueStringPointer(),
			Account: cfg.AccountID.ValueStringPointer(),
			Include: ptr("created-by"),
		}
		environment, err = GetEnvironmentByName(ctx, options, d.Client)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving environment", err.Error())
			return
		}
		if !cfg.Id.IsNull() && cfg.Id.ValueString() != environment.ID {
			resp.Diagnostics.AddError(
				"Error retrieving environment",
				fmt.Sprintf("Could not find environment with ID '%s' and name '%s'", cfg.Id.ValueString(), cfg.Name.ValueString()),
			)
			return
		}
	}

	// Update the configuration.

	cfg.Id = types.StringValue(environment.ID)
	cfg.Name = types.StringValue(environment.Name)
	cfg.Status = types.StringValue(string(environment.Status))
	cfg.RemoteBackend = types.BoolValue(environment.RemoteBackend)
	cfg.RemoteBackendOverridable = types.BoolValue(environment.RemoteBackendOverridable)
	cfg.MaskSensitiveOutput = types.BoolValue(environment.MaskSensitiveOutput)
	cfg.AccountID = types.StringValue(environment.Account.ID)

	if environment.CreatedBy != nil {
		createdBy := []userModel{*userModelFromAPI(environment.CreatedBy)}
		createdByValue, d := types.ListValueFrom(ctx, userElementType, createdBy)
		resp.Diagnostics.Append(d...)
		cfg.CreatedBy = createdByValue
	}

	policyGroups := make([]string, len(environment.PolicyGroups))
	for i, group := range environment.PolicyGroups {
		policyGroups[i] = group.ID
	}
	sort.Strings(policyGroups)
	policyGroupsValue, diags := types.ListValueFrom(ctx, types.StringType, policyGroups)
	resp.Diagnostics.Append(diags...)
	cfg.PolicyGroups = policyGroupsValue

	defaultPcfgs := make([]string, len(environment.DefaultProviderConfigurations))
	for i, pcfg := range environment.DefaultProviderConfigurations {
		defaultPcfgs[i] = pcfg.ID
	}
	sort.Strings(defaultPcfgs)
	defaultPcfgsValue, diags := types.ListValueFrom(ctx, types.StringType, defaultPcfgs)
	resp.Diagnostics.Append(diags...)
	cfg.DefaultProviderConfigurations = defaultPcfgsValue

	tags := make([]string, len(environment.Tags))
	for i, tag := range environment.Tags {
		tags[i] = tag.ID
	}
	sort.Strings(tags)
	tagsValue, diags := types.ListValueFrom(ctx, types.StringType, tags)
	resp.Diagnostics.Append(diags...)
	cfg.TagIDs = tagsValue

	var federatedEnvironments []string
	if environment.IsFederatedToAccount {
		federatedEnvironments = []string{"*"}
	} else {
		federatedEnvironments, err = getFederatedEnvironments(ctx, d.Client, environment.ID)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving federated environments", err.Error())
		}
	}
	federatedValue, diags := types.SetValueFrom(ctx, types.StringType, federatedEnvironments)
	resp.Diagnostics.Append(diags...)
	cfg.FederatedEnvironments = federatedValue

	if environment.StorageProfile != nil {
		cfg.StorageProfileID = types.StringValue(environment.StorageProfile.ID)
	}

	if environment.DefaultWorkspaceAgentPool != nil {
		cfg.DefaultWorkspaceAgentPoolID = types.StringValue(environment.DefaultWorkspaceAgentPool.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
