package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/scalr/terraform-provider-scalr/internal/client"
)

// Compile-time interface check
var _ provider.Provider = &scalrProvider{}

// New returns a function that creates a Scalr provider instance with version v.
func New(v string) func() provider.Provider {
	return func() provider.Provider {
		return &scalrProvider{
			version: v,
		}
	}
}

// scalrProviderModel describes the provider data model.
type scalrProviderModel struct {
	Hostname types.String `tfsdk:"hostname"`
	Token    types.String `tfsdk:"token"`
}

// scalrProvider implements the Terraform plugin framework Provider interface.
type scalrProvider struct {
	version string
}

func (p *scalrProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scalr"
	resp.Version = p.version
}

func (p *scalrProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: fmt.Sprintf("The Scalr hostname to connect to. Defaults to %q."+
					" Can be overridden by setting the %s environment variable.",
					client.DefaultHostname, client.HostnameEnvVar),
				MarkdownDescription: fmt.Sprintf("The Scalr hostname to connect to. Defaults to `%s`."+
					" Can be overridden by setting the `%s` environment variable.",
					client.DefaultHostname, client.HostnameEnvVar),
				Optional: true,
			},
			"token": schema.StringAttribute{
				Description: fmt.Sprintf("The token used to authenticate with Scalr."+
					" Can be overridden by setting the %s environment variable."+
					" See Scalr provider configuration at https://docs.scalr.io/docs/scalr"+
					" for information on generating a token.",
					client.TokenEnvVar),
				MarkdownDescription: fmt.Sprintf("The token used to authenticate with Scalr."+
					" Can be overridden by setting the `%s` environment variable."+
					" See [Scalr provider configuration](https://docs.scalr.io/docs/scalr)"+
					" for information on generating a token.",
					client.TokenEnvVar),
				Optional: true,
			},
		},
	}
}

func (p *scalrProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Scalr provider...")

	var cfg scalrProviderModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.Hostname.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("hostname"),
			"Unknown Scalr instance hostname",
			fmt.Sprintf(
				"The provider cannot create the Scalr API client as there is an unknown configuration value"+
					" for the Scalr instance hostname. Either target apply the source of the value first,"+
					" set the value statically in the configuration, or use the %s environment variable.",
				client.HostnameEnvVar,
			),
		)
	}
	if cfg.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Scalr API token",
			fmt.Sprintf(
				"The provider cannot create the Scalr API client as there is an unknown configuration value"+
					" for the Scalr API token. Either target apply the source of the value first,"+
					" set the value statically in the configuration, or use the %s environment variable.",
				client.TokenEnvVar,
			),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	hostname := os.Getenv(client.HostnameEnvVar)
	token := os.Getenv(client.TokenEnvVar)

	if !cfg.Hostname.IsNull() {
		hostname = cfg.Hostname.ValueString()
	}
	if hostname == "" {
		hostname = client.DefaultHostname
	}

	if !cfg.Token.IsNull() {
		token = cfg.Token.ValueString()
	}

	ctx = tflog.SetField(ctx, "scalr_hostname", hostname)

	tflog.Debug(ctx, "Creating Scalr client...")

	scalrClient, err := client.Configure(hostname, token, p.version)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Scalr API client",
			"An unexpected error occurred when creating the Scalr API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Scalr client error: "+err.Error(),
		)
		return
	}

	// Make the Scalr client available during DataSource and Resource Configure methods.
	resp.DataSourceData = scalrClient
	resp.ResourceData = scalrClient

	tflog.Info(ctx, "Scalr provider configured.")
}

func (p *scalrProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newAgentPoolTokenResource,
		newAssumeServiceAccountPolicyResource,
		newCheckovIntegrationResource,
		newEnvironmentHookResource,
		newEnvironmentResource,
		newHookResource,
		newIntegrationInfracostResource,
		newModuleNamespaceResource,
		newStorageProfileResource,
		newTagResource,
		newVariableResource,
		newWorkloadIdentityProviderResource,
		newWorkspaceResource,
	}
}

func (p *scalrProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newAssumeServiceAccountPolicyDataSource,
		newEnvironmentDataSource,
		newEnvironmentsDataSource,
		newHookDataSource,
		newIntegrationInfracostDataSource,
		newModuleNamespaceDataSource,
		newProviderConfigurationDataSource,
		newStorageProfileDataSource,
		newTagDataSource,
		newWorkloadIdentityProviderDataSource,
	}
}
