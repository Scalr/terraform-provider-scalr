package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	scalr "github.com/scalr/go-scalr"
)

const cloudCredential = "cred-suh84u5bfnjaa0g"

func TestAccEnvironment_basic(t *testing.T) {
	environment := &scalr.Environment{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentAttributes(environment, rInt),
					resource.TestCheckResourceAttr("scalr_environment.test", "name", fmt.Sprintf("test-env-%d", rInt)),
					resource.TestCheckResourceAttr("scalr_environment.test", "cost_estimation_enabled", "true"),
					resource.TestCheckResourceAttr("scalr_environment.test", "status", "Active"),
					resource.TestCheckResourceAttr("scalr_environment.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_environment.test", "cloud_credentials.%", "0"),
					resource.TestCheckResourceAttr("scalr_environment.test", "policy_groups.%", "0"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.username"),
				),
			},
		},
	})
}

func TestAccEnvironment_update(t *testing.T) {
	environment := &scalr.Environment{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentAttributes(environment, rInt),
					resource.TestCheckResourceAttr("scalr_environment.test", "name", fmt.Sprintf("test-env-%d", rInt)),
					resource.TestCheckResourceAttr("scalr_environment.test", "cost_estimation_enabled", "true"),
					resource.TestCheckResourceAttr("scalr_environment.test", "status", "Active"),
					resource.TestCheckResourceAttr("scalr_environment.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_environment.test", "cloud_credentials.0", cloudCredential),
					resource.TestCheckResourceAttr("scalr_environment.test", "policy_groups.%", "0"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.username"),
				),
			},
			{
				Config: testAccEnvironmentUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentAttributesUpdate(environment, rInt),
					resource.TestCheckResourceAttr("scalr_environment.test", "name", fmt.Sprintf("test-env-%d-patched", rInt)),
					resource.TestCheckResourceAttr("scalr_environment.test", "cost_estimation_enabled", "false"),
					resource.TestCheckResourceAttr("scalr_environment.test", "cloud_credentials.%", "0"),
				),
			},
			{
				Config:      testAccEnvironmentUpdateConfigEmptyString(rInt),
				ExpectError: regexp.MustCompile("Got error during parsing cloud credentials: 0-th value is empty"),
			},
		},
	})
}

func TestAccEnvironmentWithProviderConfigurations_update(t *testing.T) {
	environment := &scalr.Environment{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentWithProviderConfigurationsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentProviderConfigurations(environment, rInt),
				),
			},
			{
				Config: testAccEnvironmentWithProviderConfigurationsUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentProviderConfigurationsUpdate(environment, rInt),
				),
			},
		},
	})
}

func testAccCheckScalrEnvironmentDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_environment" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Environments.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Environment %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckScalrEnvironmentExists(n string, environment *scalr.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		env, err := scalrClient.Environments.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*environment = *env

		return nil
	}
}

func testAccCheckScalrEnvironmentAttributes(environment *scalr.Environment, rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if environment.Status != "Active" {
			return fmt.Errorf("Bad status: %s", environment.Status)
		}

		if environment.CostEstimationEnabled != true {
			return fmt.Errorf("Bad cost_estimation_enabled: %t", environment.CostEstimationEnabled)
		}
		if environment.Name != fmt.Sprintf("test-env-%d", rInt) {
			return fmt.Errorf("Bad name: %s", environment.Name)
		}
		if environment.Account.ID != defaultAccount {
			return fmt.Errorf("Bad account_id: %s", environment.Account.ID)
		}

		return nil
	}
}
func testAccCheckScalrEnvironmentAttributesUpdate(environment *scalr.Environment, rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if environment.CostEstimationEnabled != false {
			return fmt.Errorf("Bad cost_estimation_enabled: %t", environment.CostEstimationEnabled)
		}
		if environment.Name != fmt.Sprintf("test-env-%d-patched", rInt) {
			return fmt.Errorf("Bad name: %s", environment.Name)
		}
		return nil
	}
}

func testAccCheckScalrEnvironmentProviderConfigurations(environment *scalr.Environment, rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		if len(environment.DefaultProviderConfigurations) != 1 {
			return fmt.Errorf("Bad default provider configurations: %v", environment.DefaultProviderConfigurations)
		}
		provider_configuration, err := scalrClient.ProviderConfigurations.Read(ctx, environment.DefaultProviderConfigurations[0].ID)
		if err != nil {
			return err
		}
		if provider_configuration.ProviderName != "consul" {
			return fmt.Errorf("Bad default provider configurations: %s", provider_configuration.ProviderName)
		}
		return nil
	}
}
func testAccCheckScalrEnvironmentProviderConfigurationsUpdate(environment *scalr.Environment, rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		if len(environment.DefaultProviderConfigurations) != 1 {
			return fmt.Errorf("Bad default provider configurations: %v", environment.DefaultProviderConfigurations)
		}
		provider_configuration, err := scalrClient.ProviderConfigurations.Read(ctx, environment.DefaultProviderConfigurations[0].ID)
		if err != nil {
			return err
		}
		if provider_configuration.ProviderName != "kubernetes" {
			return fmt.Errorf("Bad default provider configurations: %s", provider_configuration.ProviderName)
		}
		return nil
	}
}

func testAccEnvironmentConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "%s"
  cost_estimation_enabled = true
  cloud_credentials = ["%s"]
}`, rInt, defaultAccount, cloudCredential)
}

func testAccEnvironmentUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%d-patched"
  account_id = "%s"
  cost_estimation_enabled = false
  cloud_credentials = []
}`, rInt, defaultAccount)
}

func testAccEnvironmentWithProviderConfigurationsConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "consul" {
  name         = "consul"
  account_id   = "%s"
  environments = ["*"]
  custom {
    provider_name = "consul"
    argument {
      name        = "config_path"
      value       = "config"
    }
  }
}

resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "%s"
  cost_estimation_enabled = true
  default_provider_configurations = ["${scalr_provider_configuration.consul.id}"]
}`, defaultAccount, rInt, defaultAccount)
}

func testAccEnvironmentWithProviderConfigurationsUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name         = "kubernetes"
  account_id   = "%s"
  environments = ["*"]
  custom {
    provider_name = "kubernetes"
    argument {
      name        = "config_path"
      value       = "config"
    }
  }
}

resource "scalr_environment" "test" {
  name       = "test-env-%d-patched"
  account_id = "%s"
  cost_estimation_enabled = false
  default_provider_configurations = ["${scalr_provider_configuration.kubernetes.id}"]
}`, defaultAccount, rInt, defaultAccount)
}

func testAccEnvironmentUpdateConfigEmptyString(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%d-patched"
  account_id = "%s"
  cost_estimation_enabled = false
  cloud_credentials = [""]
}`, rInt, defaultAccount)
}
