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
				Config: testAccScalrPorivderConfigurationConfig(rName),
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

	// loop through the resources in state, verifying each widget
	// is destroyed
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

// testAccScalrPorivderConfigurationConfig returns an configuration for an Provider Configuration with the provided name
func testAccScalrPorivderConfigurationConfig(name string) string {
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
