package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrGrafanaIntegrationDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-grafana")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrGrafanaIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      `data "scalr_grafana_integration" "test" {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
			},
			{
				Config:      `data "scalr_grafana_integration" "test" { id = "" }`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data "scalr_grafana_integration" "test" { name = "" }`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config:      `data "scalr_grafana_integration" "test" { id = "in-nonexistent" }`,
				ExpectError: regexp.MustCompile("Could not find Grafana integration with ID 'in-nonexistent'"),
			},
			{
				Config: testAccScalrGrafanaIntegrationDataSourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scalr_grafana_integration.by_id", "id", "scalr_grafana_integration.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_id", "name", name),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_id", "kind", "loki"),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_id", "endpoint", testGrafanaFailingEndpoint),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_id", "auth_type", "basic"),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_id", "username", "user"),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_id", "tenant_id", "tenant-1"),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_id", "status", "failed"),
					resource.TestCheckResourceAttrSet("data.scalr_grafana_integration.by_id", "err_message"),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_id", "account_id", defaultAccount),
					resource.TestCheckNoResourceAttr("data.scalr_grafana_integration.by_id", "secret"),
					resource.TestCheckResourceAttrPair("data.scalr_grafana_integration.by_name", "id", "scalr_grafana_integration.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_grafana_integration.by_name", "name", name),
					resource.TestCheckResourceAttrPair("data.scalr_grafana_integration.by_id_and_name", "id", "scalr_grafana_integration.test", "id"),
				),
			},
			{
				Config:      testAccScalrGrafanaIntegrationDataSourceConfig(name) + `data "scalr_grafana_integration" "missing" { name = "missing-grafana-integration" }`,
				ExpectError: regexp.MustCompile("Could not find Grafana integration"),
			},
		},
	})
}

func testAccScalrGrafanaIntegrationDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_grafana_integration" "test" {
  name      = %q
  kind      = "loki"
  endpoint  = %q
  auth_type = "basic"
  username  = "user"
  secret    = "fake-password"
  tenant_id = "tenant-1"
}

data "scalr_grafana_integration" "by_id" {
  id = scalr_grafana_integration.test.id
}

data "scalr_grafana_integration" "by_name" {
  name = scalr_grafana_integration.test.name
}

data "scalr_grafana_integration" "by_id_and_name" {
  id   = scalr_grafana_integration.test.id
  name = scalr_grafana_integration.test.name
}
`, name, testGrafanaFailingEndpoint)
}
