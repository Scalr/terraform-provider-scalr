package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccProviderConfigurationDefault_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckProviderConfigurationDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigurationDefaultBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderConfigurationDefaultExists(
						"scalr_provider_configuration_default.test",
					),
				),
			},
		},
	})
}

func TestAccProviderConfigurationDefault_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckProviderConfigurationDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigurationDefaultBasicConfig(rInt),
			},
			{
				ResourceName:      "scalr_provider_configuration_default.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckProviderConfigurationDefaultExists(
	rn string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("Not found: %s", rn)
		}

		client := testAccProvider.Meta().(*scalr.Client)

		providerConfigurationID := rs.Primary.Attributes["provider_configuration_id"]
		environmentID := rs.Primary.Attributes["environment_id"]

		environment, err := client.Environments.Read(ctx, environmentID)
		if err != nil {
			return err
		}

		for _, defaultProviderConfiguration := range environment.DefaultProviderConfigurations {
			if defaultProviderConfiguration.ID == providerConfigurationID {
				return nil
			}
		}

		return fmt.Errorf("Provider configuration default not found")
	}
}

func testAccCheckProviderConfigurationDefaultDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_provider_configuration_default" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		environment, err := scalrClient.Environments.Read(ctx, rs.Primary.Attributes["environment_id"])
		if err == nil {
			for _, defaultProviderConfiguration := range environment.DefaultProviderConfigurations {
				if defaultProviderConfiguration.ID == rs.Primary.Attributes["provider_configuration_id"] {
					return fmt.Errorf("Provider configuration default %s still exists", rs.Primary.ID)
				}
			}
		}
	}

	return nil
}

func testAccProviderConfigurationDefaultBasicConfig(rInt int) string {
	return fmt.Sprintf(`
locals {
  account_id = "%s"
}

resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = local.account_id
}

resource "scalr_provider_configuration" "test" {
  name = "test-%d"	
  account_id = local.account_id
  environments = [scalr_environment.test.id]
  custom {
	provider_name = "kubernetes"
    argument {
	  name        = "config_path"
	  value       = "~/.kube/config"
	  sensitive   = true
	  description = "A path to a kube config file."
      }
    argument {
	  name  = "host"
	  value = "my-host"
  }
   }
}

resource "scalr_provider_configuration_default" "test" {
	  provider_configuration_id = scalr_provider_configuration.test.id
	  environment_id = scalr_environment.test.id
}
`, defaultAccount, rInt, rInt)
}
