package scalr

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/scalr/go-scalr"
	providerVersion "github.com/scalr/terraform-provider-scalr/version"
)

const defaultHostname = "scalr.io"

var scalrServiceIDs = []string{"iacp.v3"}

// Config is the structure of the configuration for the Terraform CLI.
type Config struct {
	Hosts       map[string]*ConfigHost            `hcl:"host"`
	Credentials map[string]map[string]interface{} `hcl:"credentials"`
}

// ConfigHost is the structure of the "host" nested block within the CLI
// configuration, which can be used to override the default service host
// discovery behavior for a particular hostname.
type ConfigHost struct {
	Services map[string]interface{} `hcl:"services"`
}

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Scalr instance hostname without scheme. Defaults to %s.", defaultHostname),
				DefaultFunc: schema.EnvDefaultFunc("SCALR_HOSTNAME", defaultHostname),
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Scalr API token.",
				DefaultFunc: schema.EnvDefaultFunc("SCALR_TOKEN", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"scalr_access_policy":           dataSourceScalrAccessPolicy(),
			"scalr_agent_pool":              dataSourceScalrAgentPool(),
			"scalr_current_account":         dataSourceScalrCurrentAccount(),
			"scalr_current_run":             dataSourceScalrCurrentRun(),
			"scalr_endpoint":                dataSourceScalrEndpoint(),
			"scalr_environment":             dataSourceScalrEnvironment(),
			"scalr_iam_team":                dataSourceScalrIamTeam(),
			"scalr_iam_user":                dataSourceScalrIamUser(),
			"scalr_module_version":          dataSourceModuleVersion(),
			"scalr_policy_group":            dataSourceScalrPolicyGroup(),
			"scalr_provider_configuration":  dataSourceScalrProviderConfiguration(),
			"scalr_provider_configurations": dataSourceScalrProviderConfigurations(),
			"scalr_role":                    dataSourceScalrRole(),
			"scalr_service_account":         dataSourceScalrServiceAccount(),
			"scalr_tag":                     dataSourceScalrTag(),
			"scalr_variable":                dataSourceScalrVariable(),
			"scalr_variables":               dataSourceScalrVariables(),
			"scalr_vcs_provider":            dataSourceScalrVcsProvider(),
			"scalr_webhook":                 dataSourceScalrWebhook(),
			"scalr_workspace":               dataSourceScalrWorkspace(),
			"scalr_workspace_ids":           dataSourceScalrWorkspaceIDs(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"scalr_access_policy":          resourceScalrAccessPolicy(),
			"scalr_account_allowed_ips":    resourceScalrAccountAllowedIps(),
			"scalr_agent_pool":             resourceScalrAgentPool(),
			"scalr_agent_pool_token":       resourceScalrAgentPoolToken(),
			"scalr_endpoint":               resourceScalrEndpoint(),
			"scalr_environment":            resourceScalrEnvironment(),
			"scalr_iam_team":               resourceScalrIamTeam(),
			"scalr_module":                 resourceScalrModule(),
			"scalr_policy_group":           resourceScalrPolicyGroup(),
			"scalr_policy_group_linkage":   resourceScalrPolicyGroupLinkage(),
			"scalr_provider_configuration": resourceScalrProviderConfiguration(),
			"scalr_role":                   resourceScalrRole(),
			"scalr_run_trigger":            resourceScalrRunTrigger(),
			"scalr_service_account":        resourceScalrServiceAccount(),
			"scalr_tag":                    resourceScalrTag(),
			"scalr_variable":               resourceScalrVariable(),
			"scalr_vcs_provider":           resourceScalrVcsProvider(),
			"scalr_webhook":                resourceScalrWebhook(),
			"scalr_workspace":              resourceScalrWorkspace(),
			"scalr_workspace_run_schedule": resourceScalrWorkspaceRunSchedule(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Parse the hostname for comparison,
	hostname, err := svchost.ForComparison(d.Get("hostname").(string))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	providerUaString := fmt.Sprintf("terraform-provider-scalr/%s", providerVersion.ProviderVersion)

	// Get the Terraform CLI configuration.
	config := cliConfig()

	// Create a new credential source and service discovery object.
	credsSrc := credentialsSource(config)
	services := disco.NewWithCredentialsSource(credsSrc)
	services.SetUserAgent(providerUaString)
	services.Transport = logging.NewLoggingHTTPTransport(services.Transport)

	// Add any static host configurations service discovery object.
	for userHost, hostConfig := range config.Hosts {
		host, err := svchost.ForComparison(userHost)
		if err != nil {
			// ignore invalid hostnames.
			continue
		}
		services.ForceHostServices(host, hostConfig.Services)
	}

	// Discover the address.
	host, err := services.Discover(hostname)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// Get the full service address.
	var address *url.URL
	var discoErr error
	for _, scalrServiceID := range scalrServiceIDs {
		service, err := host.ServiceURL(scalrServiceID)
		if _, ok := err.(*disco.ErrVersionNotSupported); !ok && err != nil {
			return nil, diag.FromErr(err)
		}
		// If discoErr is nil we save the first error. When multiple services
		// are checked, and we found one that didn't give an error we need to
		// reset the discoErr. So if err is nil, we assign it as well.
		if discoErr == nil || err == nil {
			discoErr = err
		}
		if service != nil {
			address = service
			break
		}
	}

	// When we don't have any constraints errors, also check for discovery
	// errors before we continue.
	if discoErr != nil {
		return nil, diag.FromErr(discoErr)
	}

	// Get the token from the config.
	token := d.Get("token").(string)

	// Only try to get to the token from the credentials source if no token
	// was explicitly set in the provider configuration.
	if token == "" {
		creds, err := services.CredentialsForHost(hostname)
		if err != nil {
			log.Printf("[DEBUG] Failed to get credentials for %s: %s (ignoring)", hostname, err)
		}
		if creds != nil {
			token = creds.Token()
		}
	}

	// If we still don't have a token at this point, we return an error.
	if token == "" {
		return nil, diag.Errorf("required token could not be found")
	}

	httpClient := scalr.DefaultConfig().HTTPClient
	httpClient.Transport = logging.NewLoggingHTTPTransport(httpClient.Transport)

	headers := make(http.Header)
	headers.Add("User-Agent", providerUaString)

	// Create a new Scalr client config
	cfg := &scalr.Config{
		Address:    address.String(),
		Token:      token,
		HTTPClient: httpClient,
		Headers:    headers,
	}

	// Create a new Scalr client.
	client, err := scalr.NewClient(cfg)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	client.RetryServerErrors(true)
	return client, nil
}

// cliConfig tries to find and parse the configuration of the Terraform CLI.
// This is an optional step, so any errors are ignored.
func cliConfig() *Config {
	config := &Config{}

	// Detect the CLI config file path.
	configFilePath := os.Getenv("TERRAFORM_CONFIG")
	if configFilePath == "" {
		filePath, err := configFile()
		if err != nil {
			log.Printf("[ERROR] Error detecting default CLI config file path: %s", err)
			return config
		}
		configFilePath = filePath
	}

	// Read the CLI config file content.
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Printf("[ERROR] Error reading the CLI config file %s: %v", configFilePath, err)
		return config
	}

	// Parse the CLI config file content.
	obj, err := hcl.Parse(string(content))
	if err != nil {
		log.Printf("[ERROR] Error parsing the CLI config file %s: %v", configFilePath, err)
		return config
	}

	// Decode the CLI config file content.
	if err := hcl.DecodeObject(&config, obj); err != nil {
		log.Printf("[ERROR] Error decoding the CLI config file %s: %v", configFilePath, err)
	}

	return config
}

func credentialsSource(config *Config) auth.CredentialsSource {
	creds := auth.NoCredentials

	// Add all configured credentials to the credentials source.
	if len(config.Credentials) > 0 {
		staticTable := map[svchost.Hostname]map[string]interface{}{}
		for userHost, creds := range config.Credentials {
			host, err := svchost.ForComparison(userHost)
			if err != nil {
				// We expect the config was already validated by the time we get
				// here, so we'll just ignore invalid hostnames.
				continue
			}
			staticTable[host] = creds
		}
		creds = auth.StaticCredentialsSource(staticTable)
	}

	return creds
}

// checkConstraints checks service version constrains against our own
// version and returns rich and informational diagnostics in case any
// incompatibilities are detected.
// nolint:deadcode,unused
func checkConstraints(c *disco.Constraints) error {
	if c == nil || c.Minimum == "" || c.Maximum == "" {
		return nil
	}

	// Generate a parsable constraints string.
	excluding := ""
	if len(c.Excluding) > 0 {
		excluding = fmt.Sprintf(", != %s", strings.Join(c.Excluding, ", != "))
	}
	constStr := fmt.Sprintf(">= %s%s, <= %s", c.Minimum, excluding, c.Maximum)

	// Create the constraints to check against.
	constraints, err := version.NewConstraint(constStr)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	// Create the version to check.
	v, err := version.NewVersion(providerVersion.ProviderVersion)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	// Return if we satisfy all constraints.
	if constraints.Check(v) {
		return nil
	}

	// Find out what action (upgrade/downgrade) we should advise.
	minimum, err := version.NewVersion(c.Minimum)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	maximum, err := version.NewVersion(c.Maximum)
	if err != nil {
		return checkConstraintsWarning(err)
	}

	var excludes []*version.Version
	for _, exclude := range c.Excluding {
		v, err := version.NewVersion(exclude)
		if err != nil {
			return checkConstraintsWarning(err)
		}
		excludes = append(excludes, v)
	}

	// Sort all the excludes.
	sort.Sort(version.Collection(excludes))

	var action, toVersion string
	switch {
	case minimum.GreaterThan(v):
		action = "upgrade"
		toVersion = ">= " + minimum.String()
	case maximum.LessThan(v):
		action = "downgrade"
		toVersion = "<= " + maximum.String()
	case len(excludes) > 0:
		// Get the latest excluded version.
		action = "upgrade"
		toVersion = "> " + excludes[len(excludes)-1].String()
	}

	switch {
	case len(excludes) == 1:
		excluding = fmt.Sprintf(", excluding version %s", excludes[0].String())
	case len(excludes) > 1:
		var vs []string
		for _, v := range excludes {
			vs = append(vs, v.String())
		}
		excluding = fmt.Sprintf(", excluding versions %s", strings.Join(vs, ", "))
	default:
		excluding = ""
	}

	summary := fmt.Sprintf("Incompatible Scalr provider version v%s", v.String())
	details := fmt.Sprintf(
		"The configured Scalr installation is compatible with Scalr provider\n"+
			"versions >= %s, <= %s%s.", c.Minimum, c.Maximum, excluding,
	)

	if action != "" && toVersion != "" {
		summary = fmt.Sprintf("Please %s the Scalr provider to %s", action, toVersion)
	}

	// Return the customized and informational error message.
	return fmt.Errorf("%s\n\n%s", summary, details)
}

// nolint:unused
func checkConstraintsWarning(err error) error {
	return fmt.Errorf(
		"Failed to check version constraints: %v\n\n"+
			"Checking version constraints is considered optional, but this is an\n"+
			"unexpected error which should be reported.",
		err,
	)
}
