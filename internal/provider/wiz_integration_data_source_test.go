package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrWizIntegrationDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-wiz")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWizIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_wiz_integration test {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
			},
			{
				Config:      `data scalr_wiz_integration test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data scalr_wiz_integration test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config:      `data scalr_wiz_integration test {id = "int-nonexistent"}`,
				ExpectError: regexp.MustCompile("Could not find Wiz integration with ID 'int-nonexistent'"),
			},
			{
				Config:      `data scalr_wiz_integration test {name = "int-nonexistent"}`,
				ExpectError: regexp.MustCompile("Could not find Wiz integration with name 'int-nonexistent'"),
			},
			{
				Config: testAccScalrWizIntegrationDataSourceByIDConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_wiz_integration.test", "id",
						"scalr_wiz_integration.test", "id",
					),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "name", name),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "client_id", "test-client-id"),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "autofail", "true"),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "endpoint_mode", "govcloud"),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "default_scan_name", "nightly"),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "default_policies.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.scalr_wiz_integration.test", "default_policies.*", "Default IaC policy"),
					resource.TestCheckTypeSetElemAttr("data.scalr_wiz_integration.test", "default_policies.*", "Custom policy"),
					resource.TestCheckResourceAttrSet("data.scalr_wiz_integration.test", "status"),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "environments.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.scalr_wiz_integration.test", "environments.*", "*"),
				),
			},
			{
				Config: testAccScalrWizIntegrationDataSourceByNameConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_wiz_integration.test", "id",
						"scalr_wiz_integration.test", "id",
					),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "name", name),
				),
			},
			{
				Config: testAccScalrWizIntegrationDataSourceByIDAndNameConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_wiz_integration.test", "id",
						"scalr_wiz_integration.test", "id",
					),
					resource.TestCheckResourceAttr("data.scalr_wiz_integration.test", "name", name),
				),
			},
			{
				Config:      testAccScalrWizIntegrationDataSourceByIDAndWrongNameConfig(name),
				ExpectError: regexp.MustCompile("Could not find Wiz integration with ID"),
			},
		},
	})
}

func testAccScalrWizIntegrationDataSourceResource(name string) string {
	return fmt.Sprintf(`
resource "scalr_wiz_integration" "test" {
  name          = "%s"
  client_id     = "test-client-id"
  client_secret = "test-client-secret"

  default_policies  = ["Default IaC policy", "Custom policy"]
  default_scan_name = "nightly"
  autofail          = true
  endpoint_mode     = "govcloud"
  environments      = ["*"]
}`, name)
}

func testAccScalrWizIntegrationDataSourceByIDConfig(name string) string {
	return testAccScalrWizIntegrationDataSourceResource(name) + `

data "scalr_wiz_integration" "test" {
  id = scalr_wiz_integration.test.id
}`
}

func testAccScalrWizIntegrationDataSourceByNameConfig(name string) string {
	return testAccScalrWizIntegrationDataSourceResource(name) + `

data "scalr_wiz_integration" "test" {
  name = scalr_wiz_integration.test.name
}`
}

func testAccScalrWizIntegrationDataSourceByIDAndNameConfig(name string) string {
	return testAccScalrWizIntegrationDataSourceResource(name) + `

data "scalr_wiz_integration" "test" {
  id   = scalr_wiz_integration.test.id
  name = scalr_wiz_integration.test.name
}`
}

func testAccScalrWizIntegrationDataSourceByIDAndWrongNameConfig(name string) string {
	return testAccScalrWizIntegrationDataSourceResource(name) + `

data "scalr_wiz_integration" "test" {
  id   = scalr_wiz_integration.test.id
  name = "${scalr_wiz_integration.test.name}-other"
}`
}
