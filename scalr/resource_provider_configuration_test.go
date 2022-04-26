package scalr

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	scalr "github.com/scalr/go-scalr"
)

func TestAccProviderConfiguration_custom(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrPorivderConfigurationCustomConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.kubernetes", &providerConfiguration),
					testAccCheckProviderConfigurationCustomValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "export_shell_variables", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.provider_type", "kubernetes"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.#", "3"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.2940088933.name", "host"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.2940088933.sensitive", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.2940088933.value", "my-host"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3779616726.name", "client_certificate"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3779616726.sensitive", "true"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.3779616726.value", "-----BEGIN CERTIFICATE-----\nMIIB9TCCAWACAQAwgbgxG"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.416308637.name", "config_path"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.416308637.sensitive", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.kubernetes", "custom.0.argument.416308637.value", "~/.kube/config"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_aws(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrPorivderConfigurationAwsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.aws", &providerConfiguration),
					testAccCheckProviderConfigurationAwsValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.access_key", "my-access-key"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.aws", "aws.0.secret_key", "my-secret-key"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_google(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrPorivderConfigurationGoogleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.google", &providerConfiguration),
					testAccCheckProviderConfigurationGoogleValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "azurerm.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.project", "my-project"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.google", "google.0.credentials", "my-credentials"),
				),
			},
		},
	})
}

func TestAccProviderConfiguration_azurerm(t *testing.T) {
	var providerConfiguration scalr.ProviderConfiguration
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProviderConfigurationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrPorivderConfigurationAzurermConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationExists("scalr_provider_configuration.azurerm", &providerConfiguration),
					testAccCheckProviderConfigurationAzurermValues(&providerConfiguration, rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "name", rName),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "export_shell_variables", "false"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "aws.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "google.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.#", "1"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "custom.#", "0"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_id", "my-client-id"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.client_secret", "my-client-secret"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.subscription_id", "my-subscription-id"),
					resource.TestCheckResourceAttr("scalr_provider_configuration.azurerm", "azurerm.0.tenant_id", "my-tenant-id"),
				),
			},
		},
	})
}

func testAccCheckProviderConfigurationCustomValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderType != "kubernetes" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "kubernetes", providerConfiguration.ProviderType)
		}
		if providerConfiguration.ExportShellVariables != true {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", true, providerConfiguration.ExportShellVariables)
		}
		expectedArguments := []scalr.ProviderConfigurationParameter{
			{Key: "config_path", Sensitive: false, Value: "~/.kube/config"},
			{Key: "client_certificate", Sensitive: true, Value: ""},
			{Key: "host", Sensitive: false, Value: "my-host"},
		}
		receivedArguments := make(map[string]scalr.ProviderConfigurationParameter)
		for _, receivedArgument := range providerConfiguration.Parameters {
			receivedArguments[receivedArgument.Key] = *receivedArgument
		}
		for _, expectedArgument := range expectedArguments {
			receivedArgument, ok := receivedArguments[expectedArgument.Key]
			if !ok {
				return fmt.Errorf("argument \"%s\" not found", expectedArgument.Key)
			} else if expectedArgument.Sensitive != receivedArgument.Sensitive {
				return fmt.Errorf("argument \"%s\" bad Sensitive, expected \"%t\", got: \"%t\"", expectedArgument.Key, expectedArgument.Sensitive, receivedArgument.Sensitive)
			} else if !receivedArgument.Sensitive && expectedArgument.Value != receivedArgument.Value {
				return fmt.Errorf("argument \"%s\" bad Value, expected \"%s\", got: \"%s\"", expectedArgument.Key, expectedArgument.Value, receivedArgument.Value)
			}
		}
		return nil
	}
}

func testAccCheckProviderConfigurationAwsValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderType != "aws" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "aws", providerConfiguration.ProviderType)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.AwsAccessKey != "my-access-key" {
			return fmt.Errorf("bad aws access key, expected \"%s\", got: %#v", "my-access-key", providerConfiguration.AwsAccessKey)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationGoogleValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderType != "google" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "google", providerConfiguration.ProviderType)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.GoogleProject != "my-project" {
			return fmt.Errorf("bad google project, expected \"%s\", got: %#v", "my-project", providerConfiguration.GoogleProject)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationAzurermValues(providerConfiguration *scalr.ProviderConfiguration, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providerConfiguration.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, providerConfiguration.Name)
		}
		if providerConfiguration.ProviderType != "azurerm" {
			return fmt.Errorf("bad provider type, expected \"%s\", got: %#v", "azurerm", providerConfiguration.ProviderType)
		}
		if providerConfiguration.ExportShellVariables != false {
			return fmt.Errorf("bad export shell variables, expected \"%t\", got: %#v", false, providerConfiguration.ExportShellVariables)
		}
		if providerConfiguration.AzurermClientId != "my-client-id" {
			return fmt.Errorf("bad azurerm client id, expected \"%s\", got: %#v", "my-client-id", providerConfiguration.AzurermClientId)
		}
		if providerConfiguration.AzurermSubscriptionId != "my-subscription-id" {
			return fmt.Errorf("bad azurerm subscription id, expected \"%s\", got: %#v", "my-subscription-id", providerConfiguration.AzurermSubscriptionId)
		}
		if providerConfiguration.AzurermTenantId != "my-tenant-id" {
			return fmt.Errorf("bad azurerm tenant id, expected \"%s\", got: %#v", "my-tenant-id", providerConfiguration.AzurermTenantId)
		}
		return nil
	}
}

func testAccCheckProviderConfigurationExists(n string, providerConfiguration *scalr.ProviderConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		scalrClient := testAccProvider.Meta().(*scalr.Client)

		providerConfigurationResource, err := scalrClient.ProviderConfigurations.Read(ctx, rs.Primary.ID)

		if err != nil {
			return err
		}

		*providerConfiguration = *providerConfigurationResource

		return nil
	}
}

func testAccCheckProviderConfigurationResourceDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "example_widget" {
			continue
		}

		_, err := scalrClient.ProviderConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Provider configuraiton (%s) still exists.", rs.Primary.ID)
		}

		if !strings.Contains(err.Error(), fmt.Sprintf("ProviderConfiguration with ID '%s' not found or user unauthorized", rs.Primary.ID)) {
			return err
		}
	}

	return nil
}

func testAccScalrPorivderConfigurationCustomConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = true
  custom {
    provider_type = "kubernetes"
    argument {
      name      = "config_path"
      value     = "~/.kube/config"
      sensitive = false
    }
    argument {
      name      = "client_certificate"
      value     = "-----BEGIN CERTIFICATE-----\nMIIB9TCCAWACAQAwgbgxG"
      sensitive = true
    }
    argument {
      name      = "host"
      value     = "my-host"
    }
  }
}
`, name, defaultAccount)
}

func testAccScalrPorivderConfigurationAwsConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "aws" {
  name                   = "%s"
  account_id             = "%s"
  export_shell_variables = false
  aws {
    secret_key = "my-secret-key"
    access_key = "my-access-key"
  }
}
`, name, defaultAccount)
}

func testAccScalrPorivderConfigurationGoogleConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "%s"
  account_id = "%s"
  google {
    project     = "my-project"
    credentials = "my-credentials"
  }
}
`, name, defaultAccount)
}

func testAccScalrPorivderConfigurationAzurermConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "azurerm" {
  name       = "%s"
  account_id = "%s"
  azurerm {
    client_id       = "my-client-id"
    client_secret   = "my-client-secret"
    subscription_id = "my-subscription-id"
    tenant_id       = "my-tenant-id"
  }
}
`, name, defaultAccount)
}
