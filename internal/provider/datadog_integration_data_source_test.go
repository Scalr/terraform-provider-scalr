package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrDatadogIntegrationDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-datadog")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data "scalr_datadog_integration" "test" {}`,
				ExpectError: regexp.MustCompile("At least one of these attributes must be configured: \\[id,name]"),
			},
			{
				Config:      `data "scalr_datadog_integration" "test" { id = "" }`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data "scalr_datadog_integration" "test" { name = "" }`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config: testAccScalrDatadogIntegrationDataSourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scalr_datadog_integration.by_id", "id", "scalr_datadog_integration.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_datadog_integration.by_id", "name", name),
					resource.TestCheckResourceAttr("data.scalr_datadog_integration.by_id", "deployment_url", "https://api.datadoghq.eu"),
					resource.TestCheckResourceAttr("data.scalr_datadog_integration.by_id", "status", "failed"),
					resource.TestCheckResourceAttrSet("data.scalr_datadog_integration.by_id", "err_message"),
					resource.TestCheckResourceAttr("data.scalr_datadog_integration.by_id", "account_id", defaultAccount),
					resource.TestCheckResourceAttrPair("data.scalr_datadog_integration.by_name", "id", "scalr_datadog_integration.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_datadog_integration.by_name", "name", name),
					resource.TestCheckResourceAttrPair("data.scalr_datadog_integration.by_id_and_name", "id", "scalr_datadog_integration.test", "id"),
				),
			},
			{
				Config:      testAccScalrDatadogIntegrationDataSourceConfig(name) + `data "scalr_datadog_integration" "missing" { name = "missing-datadog-integration" }`,
				ExpectError: regexp.MustCompile("Could not find Datadog integration"),
			},
		},
	})
}

func testAccScalrDatadogIntegrationDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_datadog_integration" "test" {
  name           = %q
  api_key        = "fake-key"
  deployment_url = "https://api.datadoghq.eu"
}

data "scalr_datadog_integration" "by_id" {
  id = scalr_datadog_integration.test.id
}

data "scalr_datadog_integration" "by_name" {
  name = scalr_datadog_integration.test.name
}

data "scalr_datadog_integration" "by_id_and_name" {
  id   = scalr_datadog_integration.test.id
  name = scalr_datadog_integration.test.name
}
`, name)
}
