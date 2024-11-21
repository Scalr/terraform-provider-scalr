package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/scalr/terraform-provider-scalr/internal/client"
	"github.com/scalr/terraform-provider-scalr/version"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf("The Scalr hostname to connect to. Defaults to `%s`."+
					" Can be overridden by setting the `%s` environment variable.",
					client.DefaultHostname, client.HostnameEnvVar),
				DefaultFunc: schema.EnvDefaultFunc(client.HostnameEnvVar, client.DefaultHostname),
			},

			"token": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf("The token used to authenticate with Scalr."+
					" Can be overridden by setting the `%s` environment variable."+
					" See [Scalr provider configuration](https://docs.scalr.io/docs/scalr)"+
					" for information on generating a token.",
					client.TokenEnvVar),
				DefaultFunc: schema.EnvDefaultFunc(client.TokenEnvVar, nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"scalr_access_policy":            dataSourceScalrAccessPolicy(),
			"scalr_agent_pool":               dataSourceScalrAgentPool(),
			"scalr_current_account":          dataSourceScalrCurrentAccount(),
			"scalr_current_run":              dataSourceScalrCurrentRun(),
			"scalr_environment":              dataSourceScalrEnvironment(),
			"scalr_environments":             dataSourceScalrEnvironments(),
			"scalr_iam_team":                 dataSourceScalrIamTeam(),
			"scalr_iam_user":                 dataSourceScalrIamUser(),
			"scalr_module_version":           dataSourceModuleVersion(),
			"scalr_module_versions":          dataSourceModuleVersions(),
			"scalr_policy_group":             dataSourceScalrPolicyGroup(),
			"scalr_provider_configuration":   dataSourceScalrProviderConfiguration(),
			"scalr_provider_configurations":  dataSourceScalrProviderConfigurations(),
			"scalr_role":                     dataSourceScalrRole(),
			"scalr_service_account":          dataSourceScalrServiceAccount(),
			"scalr_tag":                      dataSourceScalrTag(),
			"scalr_variable":                 dataSourceScalrVariable(),
			"scalr_variables":                dataSourceScalrVariables(),
			"scalr_vcs_provider":             dataSourceScalrVcsProvider(),
			"scalr_webhook":                  dataSourceScalrWebhook(),
			"scalr_workspace":                dataSourceScalrWorkspace(),
			"scalr_workspace_ids":            dataSourceScalrWorkspaceIDs(),
			"scalr_workspaces":               dataSourceScalrWorkspaces(),
			"scalr_event_bridge_integration": dataSourceScalrEventBridgeIntegration(),
			"scalr_ssh_key":                  dataSourceScalrSSHKey(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"scalr_access_policy":                  resourceScalrAccessPolicy(),
			"scalr_account_allowed_ips":            resourceScalrAccountAllowedIps(),
			"scalr_agent_pool":                     resourceScalrAgentPool(),
			"scalr_agent_pool_token":               resourceScalrAgentPoolToken(),
			"scalr_environment":                    resourceScalrEnvironment(),
			"scalr_iam_team":                       resourceScalrIamTeam(),
			"scalr_module":                         resourceScalrModule(),
			"scalr_policy_group":                   resourceScalrPolicyGroup(),
			"scalr_policy_group_linkage":           resourceScalrPolicyGroupLinkage(),
			"scalr_provider_configuration":         resourceScalrProviderConfiguration(),
			"scalr_provider_configuration_default": resourceScalrProviderConfigurationDefault(),
			"scalr_role":                           resourceScalrRole(),
			"scalr_run_trigger":                    resourceScalrRunTrigger(),
			"scalr_service_account":                resourceScalrServiceAccount(),
			"scalr_service_account_token":          resourceScalrServiceAccountToken(),
			"scalr_slack_integration":              resourceScalrSlackIntegration(),
			"scalr_tag":                            resourceScalrTag(),
			"scalr_variable":                       resourceScalrVariable(),
			"scalr_vcs_provider":                   resourceScalrVcsProvider(),
			"scalr_webhook":                        resourceScalrWebhook(),
			"scalr_workspace":                      resourceScalrWorkspace(),
			"scalr_workspace_run_schedule":         resourceScalrWorkspaceRunSchedule(),
			"scalr_run_schedule_rule":              resourceScalrRunScheduleRule(),
			"scalr_event_bridge_integration":       resourceScalrEventBridgeIntegration(),
			"scalr_ssh_key":                        resourceScalrSSHKey(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	h := d.Get("hostname").(string)
	t := d.Get("token").(string)

	scalrClient, err := client.Configure(h, t, version.ProviderVersion)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return scalrClient, nil
}
