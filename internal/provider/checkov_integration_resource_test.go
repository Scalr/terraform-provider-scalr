package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/scalr/go-scalr/v2/scalr/schemas"
)

func TestCheckovIntegrationResource_Create(t *testing.T) {
	var checkovIntegration schemas.CheckovIntegration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckCheckovIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrCheckovIntegrationConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_checkov_integration.test", "id"),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "name", "test-checkov-integration"),
					resource.TestCheckResourceAttrSet("scalr_checkov_integration.test", "version"),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "external_checks_enabled", "false"),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "cli_args", "--quiet"),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "scan_mode", "configuration-files"),
					testAccCheckCheckovIntegrationIsShared("scalr_checkov_integration.test", true, &checkovIntegration),
				),
			},
		},
	})
}

func TestCheckovIntegrationResource_Create_PlanScanMode(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckCheckovIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrCheckovIntegrationConfigScanMode("plan-json"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "scan_mode", "plan-json"),
				),
			},
		},
	})
}

func TestCheckovIntegrationResource_Create_InvalidScanMode(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrCheckovIntegrationConfigScanMode("state"),
				ExpectError: regexp.MustCompile(`Attribute scan_mode value must be one of: \["configuration-files" "plan-json"]`),
			},
		},
	})
}

func TestCheckovIntegrationResource_Create_MissedVcsAttr(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrCheckovIntegrationConfigMissedVcsProvider(),
				ExpectError: regexp.MustCompile(`These attributes must be configured together: \[vcs_provider_id,vcs_repo]`),
			},
			{
				Config:      testAccScalrCheckovIntegrationConfigMissedVcsProvider(),
				ExpectError: regexp.MustCompile(`These attributes must be configured together: \[vcs_provider_id,vcs_repo]`),
			},
		},
	})
}

func TestCheckovIntegrationResource_Update(t *testing.T) {
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	var checkovIntegration schemas.CheckovIntegration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckCheckovIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrCheckovIntegrationConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "name", "test-checkov-integration"),
					testAccCheckCheckovIntegrationIsShared("scalr_checkov_integration.test", true, &checkovIntegration),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "cli_args", "--quiet"),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "scan_mode", "configuration-files"),
				),
			},
			{
				Config: testAccScalrCheckovIntegrationConfigUpdate(rNewName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "name", rNewName),
					testAccCheckCheckovIntegrationIsShared("scalr_checkov_integration.test", false, &checkovIntegration),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "cli_args", "--compact"),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "scan_mode", "plan-json"),
				),
			},
		},
	})
}

func TestCheckovIntegrationResource_ImportState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckCheckovIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrCheckovIntegrationConfig(),
			},
			{
				ResourceName:      "scalr_checkov_integration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccScalrCheckovIntegrationConfig() string {
	return `
resource "scalr_checkov_integration" "test" {
  name         = "test-checkov-integration"
  environments = ["*"]
  cli_args     = "--quiet"
}`
}

func testAccScalrCheckovIntegrationConfigScanMode(scanMode string) string {
	return fmt.Sprintf(`
resource "scalr_checkov_integration" "test" {
  name         = "test-checkov-integration"
  environments = ["*"]
  scan_mode    = "%s"
}`, scanMode)
}

func testAccScalrCheckovIntegrationConfigMissedVcsProvider() string {
	return `
resource "scalr_checkov_integration" "test" {
  name         = "test-checkov-integration"
  vcs_repo {
   identifier = "TestRepo/local"
   branch     = "main"
  }
}`
}

func testAccScalrCheckovIntegrationConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "scalr_checkov_integration" "test" {
  name         = "%s"
  environments = []
  cli_args     = "--compact"
  scan_mode    = "plan-json"
}`, name)
}

func testAccCheckCheckovIntegrationDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_checkov_integration" {
			continue
		}

		_, err := scalrClient.CheckovIntegration.GetCheckovIntegration(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Checkov integration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCheckovIntegrationIsShared(resourceName string, expectedIsShared bool, checkovIntegration *schemas.CheckovIntegration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		scalrClient := createScalrClientV2()
		ci, err := scalrClient.CheckovIntegration.GetCheckovIntegration(ctx, rs.Primary.ID, nil)
		if err != nil {
			return fmt.Errorf("Error reading Checkov integration: %s", err)
		}

		*checkovIntegration = *ci

		if ci.Attributes.IsShared != expectedIsShared {
			return fmt.Errorf("Expected IsShared to be %t, but got %t", expectedIsShared, ci.Attributes.IsShared)
		}
		return nil
	}
}
