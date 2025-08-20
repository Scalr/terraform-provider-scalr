package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccEnvironment_basic(t *testing.T) {
	environment := &scalr.Environment{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentAttributes(environment, rInt),
					resource.TestCheckResourceAttr("scalr_environment.test", "name", fmt.Sprintf("test-env-%d", rInt)),
					resource.TestCheckResourceAttr("scalr_environment.test", "remote_backend", "false"),
					resource.TestCheckResourceAttr("scalr_environment.test", "remote_backend_overridable", "false"),
					resource.TestCheckResourceAttr("scalr_environment.test", "status", "Active"),
					resource.TestCheckResourceAttr("scalr_environment.test", "account_id", defaultAccount),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentAttributes(environment, rInt),
					resource.TestCheckResourceAttr("scalr_environment.test", "name", fmt.Sprintf("test-env-%d", rInt)),
					resource.TestCheckResourceAttr("scalr_environment.test", "status", "Active"),
					resource.TestCheckResourceAttr("scalr_environment.test", "remote_backend_overridable", "false"),
					resource.TestCheckResourceAttr("scalr_environment.test", "account_id", defaultAccount),
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
					resource.TestCheckResourceAttr("scalr_environment.test", "remote_backend_overridable", "true"),
				),
			},
		},
	})
}

func TestAccEnvironmentWithProviderConfigurations_update(t *testing.T) {
	environment := &scalr.Environment{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentWithProviderConfigurationsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentProviderConfigurations(environment),
				),
			},
			{
				Config: testAccEnvironmentWithProviderConfigurationsUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentProviderConfigurationsUpdate(environment),
				),
			},
			{
				Config: testAccEnvironmentWithProviderConfigurationsConfigRemovedDefault(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					testAccCheckScalrEnvironmentProviderConfigurationsDefaultRemoved(environment),
				),
			},
		},
	})
}

func TestAccEnvironment_UpgradeFromSDK(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"scalr": {
						Source:            "registry.scalr.io/scalr/scalr",
						VersionConstraint: "<=3.0.0",
					},
				},
				Config: testAccEnvironmentImportConfig(rInt),
				Check:  resource.TestCheckResourceAttrSet("scalr_environment.test", "id"),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(t),
				Config:                   testAccEnvironmentConfig(rInt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccEnvironment_Federated(t *testing.T) {
	environment := &scalr.Environment{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentWithFederatedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					resource.TestCheckResourceAttr(
						"scalr_environment.test", "federated_environments.#", "2"),
					testAccCheckScalrEnvironmentFederation("scalr_environment.test", false),
				),
			},
			{
				Config: testAccEnvironmentWithFederatedUpdatedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrEnvironmentExists("scalr_environment.test", environment),
					resource.TestCheckResourceAttr(
						"scalr_environment.test", "federated_environments.#", "1"),
					testAccCheckScalrEnvironmentFederation("scalr_environment.test", true),
				),
			},
		},
	})
}

func testAccCheckScalrEnvironmentDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

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
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

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

func testAccCheckScalrEnvironmentFederation(
	n string, isFederated bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

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

		if env.IsFederatedToAccount != isFederated {
			return fmt.Errorf("Expected IsFederatedToAccount %t, got %t", isFederated, env.IsFederatedToAccount)
		}

		return nil
	}
}

func testAccCheckScalrEnvironmentAttributes(environment *scalr.Environment, rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if environment.Status != "Active" {
			return fmt.Errorf("Bad status: %s", environment.Status)
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
		if environment.Name != fmt.Sprintf("test-env-%d-patched", rInt) {
			return fmt.Errorf("Bad name: %s", environment.Name)
		}
		return nil
	}
}

func testAccCheckScalrEnvironmentProviderConfigurations(environment *scalr.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		if len(environment.DefaultProviderConfigurations) != 1 {
			return fmt.Errorf("Bad default provider configurations: %v", environment.DefaultProviderConfigurations)
		}
		providerConfiguration, err := scalrClient.ProviderConfigurations.Read(ctx, environment.DefaultProviderConfigurations[0].ID)
		if err != nil {
			return err
		}
		if providerConfiguration.ProviderName != "consul" {
			return fmt.Errorf("Bad default provider configurations: %s", providerConfiguration.ProviderName)
		}
		return nil
	}
}
func testAccCheckScalrEnvironmentProviderConfigurationsUpdate(environment *scalr.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		if len(environment.DefaultProviderConfigurations) != 1 {
			return fmt.Errorf("Bad default provider configurations: %v", environment.DefaultProviderConfigurations)
		}
		providerConfiguration, err := scalrClient.ProviderConfigurations.Read(ctx, environment.DefaultProviderConfigurations[0].ID)
		if err != nil {
			return err
		}
		if providerConfiguration.ProviderName != "kubernetes" {
			return fmt.Errorf("Bad default provider configurations: %s", providerConfiguration.ProviderName)
		}
		return nil
	}
}

func testAccCheckScalrEnvironmentProviderConfigurationsDefaultRemoved(environment *scalr.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(environment.DefaultProviderConfigurations) != 0 {
			return fmt.Errorf("Bad default provider configurations: %v", environment.DefaultProviderConfigurations)
		}
		return nil
	}
}

func testAccEnvironmentConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name                       = "test-env-%d"
  account_id                 = "%s"
  remote_backend             = false
  remote_backend_overridable = false
}`, rInt, defaultAccount)
}

func testAccEnvironmentUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name                       = "test-env-%d-patched"
  remote_backend_overridable = true
  account_id                 = "%s"
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
  default_provider_configurations = ["${scalr_provider_configuration.kubernetes.id}"]
}`, defaultAccount, rInt, defaultAccount)
}

func testAccEnvironmentWithProviderConfigurationsConfigRemovedDefault(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_workspace" "test" {
  name                  = "workspace-monorepo"
  environment_id 		= scalr_environment.test.id
  working_directory     = "/db"
}

resource "scalr_environment" "test" {
  name       = "test-env-%d-patched"
  account_id = "%s"
}`, rInt, defaultAccount)
}

func testAccEnvironmentWithFederatedConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "federated1" {
  name       = "federated1-%[1]d"
  account_id = "%[2]s"
}

resource "scalr_environment" "federated2" {
  name       = "federated2-%[1]d"
  account_id = "%[2]s"
}

resource "scalr_environment" "test" {
  name       = "test-env-%[1]d"
  account_id = "%[2]s"
  federated_environments = [ scalr_environment.federated1.id, scalr_environment.federated2.id ]
}`, rInt, defaultAccount)
}

func testAccEnvironmentWithFederatedUpdatedConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "federated1" {
  name       = "federated1-%[1]d"
  account_id = "%[2]s"
}

resource "scalr_environment" "federated2" {
  name       = "federated2-%[1]d"
  account_id = "%[2]s"
}

resource "scalr_environment" "test" {
  name       = "test-env-%[1]d"
  account_id = "%[2]s"
  federated_environments = [ "*" ]
}`, rInt, defaultAccount)
}

func testAccEnvironmentImportConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name                       = "test-env-%d"
  account_id                 = "%s"
}`, rInt, defaultAccount)
}
