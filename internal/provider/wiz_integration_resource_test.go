package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccScalrWizIntegrationResource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-wiz")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWizIntegrationDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWizIntegrationBasic(name),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrWizIntegrationExists("scalr_wiz_integration.test"),
						resource.TestCheckResourceAttrSet("scalr_wiz_integration.test", "id"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "name", name),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "client_id", "test-client-id"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "client_secret", "test-client-secret"),
						// Defaults applied by the provider.
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "autofail", "false"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "endpoint_mode", "commercial"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "environments.#", "0"),
					),
				},
				{
					// The API never returns client_secret; make sure that does not
					// show up as drift.
					Config: testAccScalrWizIntegrationBasic(name),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectEmptyPlan(),
						},
					},
				},
			},
		},
	)
}

func TestAccScalrWizIntegrationResource_update(t *testing.T) {
	name := acctest.RandomWithPrefix("test-wiz")
	newName := acctest.RandomWithPrefix("test-wiz")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWizIntegrationDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWizIntegrationBasic(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "name", name),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "autofail", "false"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "endpoint_mode", "commercial"),
						resource.TestCheckNoResourceAttr("scalr_wiz_integration.test", "default_scan_name"),
					),
				},
				{
					Config: testAccScalrWizIntegrationUpdated(newName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrWizIntegrationExists("scalr_wiz_integration.test"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "name", newName),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "client_id", "updated-client-id"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "client_secret", "updated-client-secret"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "autofail", "true"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "endpoint_mode", "govcloud"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "default_scan_name", "nightly"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "default_policies.#", "2"),
						resource.TestCheckTypeSetElemAttr("scalr_wiz_integration.test", "default_policies.*", "Default IaC policy"),
						resource.TestCheckTypeSetElemAttr("scalr_wiz_integration.test", "default_policies.*", "Custom policy"),
					),
				},
			},
		},
	)
}

func TestAccScalrWizIntegrationResource_shared(t *testing.T) {
	name := acctest.RandomWithPrefix("test-wiz")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWizIntegrationDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWizIntegrationShared(name),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrWizIntegrationExists("scalr_wiz_integration.test"),
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "environments.#", "1"),
						resource.TestCheckTypeSetElemAttr("scalr_wiz_integration.test", "environments.*", "*"),
					),
				},
				{
					// Narrowing from all environments down to a single one.
					Config: testAccScalrWizIntegrationSingleEnvironment(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("scalr_wiz_integration.test", "environments.#", "1"),
						resource.TestCheckResourceAttrPair(
							"scalr_wiz_integration.test", "environments.0",
							"scalr_environment.test", "id",
						),
					),
				},
			},
		},
	)
}

func TestAccScalrWizIntegrationResource_invalidEndpointMode(t *testing.T) {
	name := acctest.RandomWithPrefix("test-wiz")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			Steps: []resource.TestStep{
				{
					Config:      testAccScalrWizIntegrationInvalidEndpointMode(name),
					ExpectError: regexp.MustCompile(`Attribute endpoint_mode value must be one of`),
				},
			},
		},
	)
}

func TestAccScalrWizIntegrationResource_import(t *testing.T) {
	name := acctest.RandomWithPrefix("test-wiz")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWizIntegrationDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWizIntegrationBasic(name),
				},
				{
					ResourceName:      "scalr_wiz_integration.test",
					ImportState:       true,
					ImportStateVerify: true,
					// The API never returns the secret, so it cannot be imported.
					ImportStateVerifyIgnore: []string{"client_secret"},
				},
			},
		},
	)
}

func testAccCheckScalrWizIntegrationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		scalrClient := createScalrClientV2()

		_, err := scalrClient.WizIntegration.GetWizIntegration(ctx, rs.Primary.ID, nil)
		if err != nil {
			return fmt.Errorf("error reading Wiz integration %s: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckScalrWizIntegrationDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_wiz_integration" {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}
		_, err := scalrClient.WizIntegration.GetWizIntegration(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Wiz integration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrWizIntegrationBasic(name string) string {
	return fmt.Sprintf(`
resource "scalr_wiz_integration" "test" {
  name          = "%s"
  client_id     = "test-client-id"
  client_secret = "test-client-secret"
}`, name)
}

func testAccScalrWizIntegrationUpdated(name string) string {
	return fmt.Sprintf(`
resource "scalr_wiz_integration" "test" {
  name          = "%s"
  client_id     = "updated-client-id"
  client_secret = "updated-client-secret"

  default_policies  = ["Default IaC policy", "Custom policy"]
  default_scan_name = "nightly"
  autofail          = true
  endpoint_mode     = "govcloud"
}`, name)
}

func testAccScalrWizIntegrationShared(name string) string {
	return fmt.Sprintf(`
resource "scalr_wiz_integration" "test" {
  name          = "%s"
  client_id     = "test-client-id"
  client_secret = "test-client-secret"
  environments  = ["*"]
}`, name)
}

func testAccScalrWizIntegrationSingleEnvironment(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name = "%s-env"
}

resource "scalr_wiz_integration" "test" {
  name          = "%s"
  client_id     = "test-client-id"
  client_secret = "test-client-secret"
  environments  = [scalr_environment.test.id]
}`, name, name)
}

func testAccScalrWizIntegrationInvalidEndpointMode(name string) string {
	return fmt.Sprintf(`
resource "scalr_wiz_integration" "test" {
  name          = "%s"
  client_id     = "test-client-id"
  client_secret = "test-client-secret"
  endpoint_mode = "invalid"
}`, name)
}
