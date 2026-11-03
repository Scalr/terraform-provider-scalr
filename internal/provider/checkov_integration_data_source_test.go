package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrCheckovIntegrationDataSource_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_checkov_integration test {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
			},
			{
				Config:      `data scalr_checkov_integration test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data scalr_checkov_integration test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config:      `data scalr_checkov_integration test {id = "in-nonexistent"}`,
				ExpectError: regexp.MustCompile("Could not find Checkov integration with ID 'in-nonexistent'"),
			},
			{
				Config:      `data scalr_checkov_integration test {name = "nonexistent"}`,
				ExpectError: regexp.MustCompile("Could not find Checkov integration with name 'nonexistent'"),
			},
		},
	})
}

func TestAccScalrCheckovIntegrationDataSource_basic(t *testing.T) {
	res := testAccScalrCheckovIntegrationConfigScanMode("plan")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckCheckovIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: res + `
data "scalr_checkov_integration" "test" {
  id = scalr_checkov_integration.test.id
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_checkov_integration.test", "id",
						"scalr_checkov_integration.test", "id",
					),
					resource.TestCheckResourceAttr("data.scalr_checkov_integration.test", "name", "test-checkov-integration"),
					resource.TestCheckResourceAttrSet("data.scalr_checkov_integration.test", "version"),
					resource.TestCheckResourceAttr("data.scalr_checkov_integration.test", "scan_mode", "plan"),
					resource.TestCheckResourceAttr("data.scalr_checkov_integration.test", "external_checks_enabled", "false"),
					resource.TestCheckResourceAttr("data.scalr_checkov_integration.test", "status", "active"),
					resource.TestCheckResourceAttr("data.scalr_checkov_integration.test", "environments.#", "1"),
					resource.TestCheckResourceAttr("data.scalr_checkov_integration.test", "environments.0", "*"),
				),
			},
			{
				Config: res + `
data "scalr_checkov_integration" "test" {
  name = "test-checkov-integration"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_checkov_integration.test", "id",
						"scalr_checkov_integration.test", "id",
					),
					resource.TestCheckResourceAttr("data.scalr_checkov_integration.test", "scan_mode", "plan"),
				),
			},
			{
				Config: res + `
data "scalr_checkov_integration" "test" {
  id   = scalr_checkov_integration.test.id
  name = "wrong-name"
}`,
				ExpectError: regexp.MustCompile(`Could not find Checkov integration with ID '[^']+', name\s+'wrong-name'`),
			},
		},
	})
}
