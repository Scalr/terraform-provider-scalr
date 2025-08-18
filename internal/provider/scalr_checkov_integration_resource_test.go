package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
	"regexp"
	"testing"
)

func TestCheckovIntegrationResource_Create(t *testing.T) {
	var checkovIntegration scalr.CheckovIntegration

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
					testAccCheckCheckovIntegrationIsShared("scalr_checkov_integration.test", true, &checkovIntegration),
				),
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

	var checkovIntegration scalr.CheckovIntegration

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
				),
			},
			{
				Config: testAccScalrCheckovIntegrationConfigUpdate(rNewName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "name", rNewName),
					testAccCheckCheckovIntegrationIsShared("scalr_checkov_integration.test", false, &checkovIntegration),
					resource.TestCheckResourceAttr("scalr_checkov_integration.test", "cli_args", "--compact"),
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
	return fmt.Sprintf(`
resource "scalr_checkov_integration" "test" {
  name         = "test-checkov-integration"
  environments = ["*"]
  cli_args     = "--quiet"
}`)
}

func testAccScalrCheckovIntegrationConfigMissedVcsProvider() string {
	return fmt.Sprintf(`
resource "scalr_checkov_integration" "test" {
  name         = "test-checkov-integration"
  vcs_repo {
   identifier = "TestRepo/local"
   branch     = "main"
  }
}`)
}

func testAccScalrCheckovIntegrationConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "scalr_checkov_integration" "test" {
  name         = "%s"
  environments = []
  cli_args     = "--compact"
}`, name)
}

func testAccCheckCheckovIntegrationDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_checkov_integration" {
			continue
		}

		_, err := scalrClient.CheckovIntegrations.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Checkov integration (%s) still exists", rs.Primary.ID)
		}

		if !errors.Is(err, scalr.ErrResourceNotFound) {
			return err
		}
	}

	return nil
}

func testAccCheckCheckovIntegrationIsShared(resourceName string, expectedIsShared bool, CheckovIntegration *scalr.CheckovIntegration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)
		readKey, err := scalrClient.CheckovIntegrations.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error reading Infracost Integration: %s", err)
		}

		*CheckovIntegration = *readKey

		if readKey.IsShared != expectedIsShared {
			return fmt.Errorf("Expected IsShared to be %t, but got %t", expectedIsShared, readKey.IsShared)
		}
		return nil
	}
}
