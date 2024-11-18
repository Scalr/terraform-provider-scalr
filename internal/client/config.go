package client

import (
	"log"
	"os"

	"github.com/hashicorp/hcl"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
)

const (
	DefaultHostname = "scalr.io"
	HostnameEnvVar  = "SCALR_HOSTNAME"
	TokenEnvVar     = "SCALR_TOKEN"
)

// config is the structure of the configuration for the Terraform CLI.
type config struct {
	Hosts       map[string]*configHost            `hcl:"host"`
	Credentials map[string]map[string]interface{} `hcl:"credentials"`
}

// configHost is the structure of the "host" nested block within the CLI
// configuration, which can be used to override the default service host
// discovery behavior for a particular hostname.
type configHost struct {
	Services map[string]interface{} `hcl:"services"`
}

// CliConfig tries to find and parse the configuration of the Terraform CLI.
// This is an optional step, so any errors are ignored.
func cliConfig() *config {
	combinedConfig := &config{}

	// Main CLI config file; might contain manually-entered credentials, and/or
	// some host service discovery objects. Location is configurable via
	// environment variables.
	mainConfig := readCliConfigFile(locateConfigFile())

	// Credentials file; might contain credentials auto-configured by terraform
	// login. Location isn't configurable.
	var credentialsConfig *config
	credentialsFilePath, err := credentialsFile()
	if err != nil {
		log.Printf("[ERROR] Error detecting default credentials file path: %s", err)
		credentialsConfig = &config{}
	} else {
		credentialsConfig = readCliConfigFile(credentialsFilePath)
	}

	// Use host service discovery configs from main config file.
	combinedConfig.Hosts = mainConfig.Hosts

	// Combine both sets of credentials. Per Terraform's own behavior, the main
	// config file overrides the credentials file if they have any overlapping
	// hostnames.
	combinedConfig.Credentials = credentialsConfig.Credentials
	if combinedConfig.Credentials == nil {
		combinedConfig.Credentials = make(map[string]map[string]interface{})
	}
	for host, creds := range mainConfig.Credentials {
		combinedConfig.Credentials[host] = creds
	}

	return combinedConfig
}

func credentialsSource(config *config) auth.CredentialsSource {
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

func readCliConfigFile(configFilePath string) *config {
	config := &config{}

	if configFilePath == "" {
		return config
	}

	// Read the CLI config file content.
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Printf("[ERROR] Error reading CLI config or credentials file %s: %v", configFilePath, err)
		return config
	}

	// Parse the CLI config file content.
	obj, err := hcl.Parse(string(content))
	if err != nil {
		log.Printf("[ERROR] Error parsing CLI config or credentials file %s: %v", configFilePath, err)
		return config
	}

	// Decode the CLI config file content.
	if err := hcl.DecodeObject(config, obj); err != nil {
		log.Printf("[ERROR] Error decoding CLI config or credentials file %s: %v", configFilePath, err)
	}

	return config
}

func locateConfigFile() string {
	// To find the main CLI config file, follow Terraform's own logic: try
	// TF_CLI_CONFIG_FILE, then try TERRAFORM_CONFIG, then try the default
	// location.

	if os.Getenv("TF_CLI_CONFIG_FILE") != "" {
		return os.Getenv("TF_CLI_CONFIG_FILE")
	}

	if os.Getenv("TERRAFORM_CONFIG") != "" {
		return os.Getenv("TERRAFORM_CONFIG")
	}
	filePath, err := configFile()
	if err != nil {
		log.Printf("[ERROR] Error detecting default CLI config file path: %s", err)
		return ""
	}

	return filePath
}
