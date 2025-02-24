package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrIntegrationInfracostDataSource_basic(t *testing.T) {
	apiKey := os.Getenv("TEST_INFRACOST_API_KEY")
	if len(apiKey) == 0 {
		t.Skip("Please set TEST_INFRACOST_API_KEY to run this test.")
	}
	integrationInfracostName := acctest.RandomWithPrefix("test-integration-infracost")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_integration_infracost test {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
			},
			{
				Config:      `data scalr_integration_infracost test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data scalr_integration_infracost test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config: testAccScalrIntegrationInfracostDataSourceByIDConfig(integrationInfracostName, apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_integration_infracost.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_integration_infracost.test", "name", integrationInfracostName),
				),
			},
			{
				Config: testAccScalrIntegrationInfracostDataSourceByNameConfig(integrationInfracostName, apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_integration_infracost.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_integration_infracost.test", "name", integrationInfracostName),
				),
			},
			{
				Config: testAccScalrIntegrationInfracostDataSourceByIDAndNameConfig(integrationInfracostName, apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_integration_infracost.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_integration_infracost.test", "name", integrationInfracostName),
				),
				Destroy: true,
			},
		},
	})
}

func testAccScalrIntegrationInfracostDataSourceByIDConfig(name string, apiKey string) string {
	return fmt.Sprintf(`
resource scalr_integration_infracost test {
  name    = "%[1]s"
  api_key = "%[2]s"
}

data scalr_integration_infracost test {
  id         = scalr_integration_infracost.test.id
}`, name, apiKey)
}

func testAccScalrIntegrationInfracostDataSourceByNameConfig(name string, apiKey string) string {
	return fmt.Sprintf(`
resource scalr_integration_infracost test {
  name    = "%[1]s"
  api_key = "%[2]s"
}

data scalr_integration_infracost test {
  name       = scalr_integration_infracost.test.name
}`, name, apiKey)
}

func testAccScalrIntegrationInfracostDataSourceByIDAndNameConfig(name string, apiKey string) string {
	return fmt.Sprintf(`
resource scalr_integration_infracost test {
  name    = "%[1]s"
  api_key = "%[2]s"
}

data scalr_integration_infracost test {
  id   = scalr_integration_infracost.test.id
  name = scalr_integration_infracost.test.name
}`, name, apiKey)
}
