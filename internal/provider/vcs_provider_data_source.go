package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
	"github.com/scalr/go-scalr/v2/scalr/ops/vcs_provider"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource              = &vcsProviderDataSource{}
	_ datasource.DataSourceWithConfigure = &vcsProviderDataSource{}
)

func newVcsProviderDataSource() datasource.DataSource {
	return &vcsProviderDataSource{}
}

// vcsProviderDataSource defines the data source implementation.
type vcsProviderDataSource struct {
	framework.DataSourceWithScalrClient
}

// vcsProviderDataSourceModel describes the data source data model.
type vcsProviderDataSourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	VcsType            types.String `tfsdk:"vcs_type"`
	Url                types.String `tfsdk:"url"`
	EnvironmentID      types.String `tfsdk:"environment_id"`
	AgentPoolID        types.String `tfsdk:"agent_pool_id"`
	Environments       types.Set    `tfsdk:"environments"`
	DraftPrRunsEnabled types.Bool   `tfsdk:"draft_pr_runs_enabled"`
	AccountID          types.String `tfsdk:"account_id"`
}

func (d *vcsProviderDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_vcs_provider"
}

func (d *vcsProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about VCS provider.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the VCS provider.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the VCS provider.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"vcs_type": schema.StringAttribute{
				MarkdownDescription: "The type of the VCS provider. For example, `github`.",
				Computed:            true,
				Optional:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL to the VCS provider installation.",
				Computed:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the environment the VCS provider has to be linked to, in the format `env-<RANDOM STRING>`.",
				Optional:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"agent_pool_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the agent pool to connect Scalr to self-hosted VCS provider, in the format `apool-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of the identifiers of the environments the VCS provider is linked to.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"draft_pr_runs_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the draft pull-request runs are enabled for this VCS provider.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage: "Setting this attribute is deprecated." +
					" It is no longer in use and will become read-only in the next major version of the provider.",
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Scalr account.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage: "Setting this attribute is deprecated." +
					" It is no longer in use and will become read-only in the next major version of the provider.",
			},
		},
	}
}

func (d *vcsProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg vcsProviderDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := vcs_provider.ListVcsProvidersOptions{
		Filter: map[string]string{},
	}
	if !cfg.Id.IsNull() {
		opts.Filter["vcs-provider"] = cfg.Id.ValueString()
	}
	if !cfg.Name.IsNull() {
		opts.Filter["name"] = cfg.Name.ValueString()
	}
	if !cfg.VcsType.IsNull() {
		opts.Filter["vcs-type"] = cfg.VcsType.ValueString()
	}
	if !cfg.EnvironmentID.IsNull() {
		opts.Filter["environment"] = cfg.EnvironmentID.ValueString()
	}
	if !cfg.AgentPoolID.IsNull() {
		opts.Filter["agent-pool"] = cfg.AgentPoolID.ValueString()
	}

	vcsProvider, err := getVcsProvider(ctx, d.ClientV2, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving VCS provider", err.Error())
		return
	}

	cfg.Id = types.StringValue(vcsProvider.ID)
	cfg.Name = types.StringValue(vcsProvider.Attributes.Name)
	cfg.VcsType = types.StringValue(string(vcsProvider.Attributes.VcsType))
	cfg.DraftPrRunsEnabled = types.BoolValue(vcsProvider.Attributes.DraftPrRunsEnabled)

	if vcsProvider.Attributes.Url != nil {
		cfg.Url = types.StringValue(*vcsProvider.Attributes.Url)
	}

	if vcsProvider.Relationships.AgentPool != nil {
		cfg.AgentPoolID = types.StringValue(vcsProvider.Relationships.AgentPool.ID)
	}

	if vcsProvider.Relationships.Account != nil {
		cfg.AccountID = types.StringValue(vcsProvider.Relationships.Account.ID)
	}

	envs := make([]string, len(vcsProvider.Relationships.Environments))
	for i, env := range vcsProvider.Relationships.Environments {
		envs[i] = env.ID
	}
	envsValue, diags := types.SetValueFrom(ctx, types.StringType, envs)
	diags.Append(diags...)
	cfg.Environments = envsValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

func getVcsProvider(
	ctx context.Context,
	client *scalrV2.Client,
	listOpts *vcs_provider.ListVcsProvidersOptions,
) (*schemas.VcsProvider, error) {
	vcsProviders, err := client.VcsProvider.ListVcsProviders(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	if len(vcsProviders) > 1 {
		return nil, fmt.Errorf(
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
	}

	if len(vcsProviders) == 0 {
		// If `name` filter was used to retrieve a VCS provider,
		// fallback to using the `query` parameter and try once more.
		// This is to keep the backward compatibility with the previous behavior,
		// where it was possible to do a partial match on the VCS provider name.
		if listOpts.Filter["name"] != "" {
			listOpts.Query = listOpts.Filter["name"]
			delete(listOpts.Filter, "name")
			return getVcsProvider(ctx, client, listOpts)
		}
		return nil, fmt.Errorf("Could not find VCS provider matching you query.")
	}

	return vcsProviders[0], nil
}
