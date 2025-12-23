package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestFederatedEnvironmentsResource_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testFederatedEnvironmentsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckScalrFederatedEnvironmentsExists("scalr_federated_environments.test"),
					testCheckScalrFederatedEnvironmentsExists("scalr_federated_environments.test2"),
					resource.TestCheckResourceAttrSet("scalr_federated_environments.test", "federated_environments.0"),
					resource.TestCheckResourceAttrSet("scalr_federated_environments.test", "environment_id"),
					resource.TestCheckResourceAttrSet("scalr_federated_environments.test2", "federated_environments.0"),
					resource.TestCheckResourceAttrSet("scalr_federated_environments.test2", "environment_id"),
				),
			},
		},
	})
}

func TestFederatedEnvironmentsResource_update(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testFederatedEnvironmentsUpdateConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testCheckScalrFederatedEnvironmentsExists("scalr_federated_environments.test"),
					resource.TestCheckResourceAttr("scalr_federated_environments.test", "federated_environments.#", "1"),
				),
			},
			{
				Config: testFederatedEnvironmentsUpdateConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testCheckScalrFederatedEnvironmentsExists("scalr_federated_environments.test"),
					resource.TestCheckResourceAttr("scalr_federated_environments.test", "federated_environments.#", "2"),
				),
			},
			{
				Config: testFederatedEnvironmentsUpdateConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testCheckScalrFederatedEnvironmentsExists("scalr_federated_environments.test"),
					resource.TestCheckResourceAttr("scalr_federated_environments.test", "federated_environments.#", "1"),
				),
			},
		},
	})
}

func TestFederatedEnvironmentsResource_shared(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testFederatedEnvironmentsToAccConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckScalrEnvironmentSharedToAccount("scalr_federated_environments.test", true),
					resource.TestCheckResourceAttr("scalr_federated_environments.test", "federated_environments.0", "*"),
					resource.TestCheckResourceAttrSet("scalr_federated_environments.test", "environment_id"),
				),
			},
		},
	})
}

func testCheckScalrFederatedEnvironmentsExists(resId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.Attributes["environment_id"] == "" {
			return errNoInstanceId
		}

		p, err := scalrClient.FederatedEnvironments.List(ctx, rs.Primary.Attributes["environment_id"], scalr.ListOptions{})
		if err != nil {
			return err
		}

		if len(p.Items) == 0 {
			return fmt.Errorf("No federated environments found for environment %s", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckScalrEnvironmentSharedToAccount(resId string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.Attributes["environment_id"] == "" {
			return errNoInstanceId
		}

		env, err := scalrClient.Environments.Read(ctx, rs.Primary.Attributes["environment_id"])
		if err != nil {
			return err
		}

		if env.IsFederatedToAccount != expected {
			return fmt.Errorf("Expected IsFederatedToAccount to be %v, got %v", expected, env.IsFederatedToAccount)
		}

		return nil
	}
}

func testFederatedEnvironmentsConfig() string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "environment-test"
  account_id = "%s"
}

resource "scalr_environment" "test2" {
  name       = "environment-test2"
  account_id = "%s"
}

resource "scalr_federated_environments" "test" {
  environment_id = scalr_environment.test.id
  federated_environments = [scalr_environment.test2.id]
}

resource "scalr_federated_environments" "test2" {
  environment_id = scalr_environment.test2.id
  federated_environments = [scalr_environment.test.id]
}
`, defaultAccount, defaultAccount)
}

func testFederatedEnvironmentsUpdateConfig(includeThird bool) string {
	federatedEnvs := "[scalr_environment.test2.id]"
	if includeThird {
		federatedEnvs = "[scalr_environment.test2.id, scalr_environment.test3.id]"
	}

	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "env-federated-update-test"
  account_id = "%s"
}

resource "scalr_environment" "test2" {
  name       = "env-federated-update-test2"
  account_id = "%s"
}

resource "scalr_environment" "test3" {
  name       = "env-federated-update-test3"
  account_id = "%s"
}

resource "scalr_federated_environments" "test" {
  environment_id = scalr_environment.test.id
  federated_environments = %s
}
`, defaultAccount, defaultAccount, defaultAccount, federatedEnvs)
}

func testFederatedEnvironmentsToAccConfig() string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "environment-test"
  account_id = "%s"
}

resource "scalr_federated_environments" "test" {
  environment_id = scalr_environment.test.id
  federated_environments = ["*"]
}
`, defaultAccount)
}
