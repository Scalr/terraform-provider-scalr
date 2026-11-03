package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrSplunkIntegrationDataSource_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data "scalr_splunk_integration" "test" {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
			},
			{
				Config:      `data "scalr_splunk_integration" "test" { id = "" }`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data "scalr_splunk_integration" "test" { name = "" }`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config:      `data "scalr_splunk_integration" "test" { id = "in-nonexistent" }`,
				ExpectError: regexp.MustCompile("Could not find Splunk integration with ID 'in-nonexistent'"),
			},
		},
	})
}

func TestAccScalrSplunkIntegrationDataSource_basic(t *testing.T) {
	endpoint := os.Getenv("TEST_SPLUNK_ENDPOINT")
	token := os.Getenv("TEST_SPLUNK_TOKEN")
	index := os.Getenv("TEST_SPLUNK_INDEX")
	if endpoint == "" || token == "" || index == "" {
		t.Skip("Please set TEST_SPLUNK_ENDPOINT, TEST_SPLUNK_TOKEN and TEST_SPLUNK_INDEX to run this test")
	}
	name := acctest.RandomWithPrefix("test-splunk")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrSplunkIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrSplunkIntegrationDataSourceConfig(name, endpoint, token, index),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scalr_splunk_integration.by_id", "id", "scalr_splunk_integration.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_splunk_integration.by_id", "name", name),
					resource.TestCheckResourceAttr("data.scalr_splunk_integration.by_id", "endpoint", endpoint),
					resource.TestCheckResourceAttr("data.scalr_splunk_integration.by_id", "index", index),
					resource.TestCheckResourceAttr("data.scalr_splunk_integration.by_id", "source", "scalr-test"),
					resource.TestCheckNoResourceAttr("data.scalr_splunk_integration.by_id", "sourcetype"),
					resource.TestCheckResourceAttr("data.scalr_splunk_integration.by_id", "status", "active"),
					resource.TestCheckNoResourceAttr("data.scalr_splunk_integration.by_id", "err_message"),
					resource.TestCheckResourceAttr("data.scalr_splunk_integration.by_id", "account_id", defaultAccount),
					resource.TestCheckNoResourceAttr("data.scalr_splunk_integration.by_id", "token"),
					resource.TestCheckResourceAttrPair("data.scalr_splunk_integration.by_name", "id", "scalr_splunk_integration.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_splunk_integration.by_name", "name", name),
					resource.TestCheckResourceAttrPair("data.scalr_splunk_integration.by_id_and_name", "id", "scalr_splunk_integration.test", "id"),
				),
			},
			{
				Config:      testAccScalrSplunkIntegrationDataSourceConfig(name, endpoint, token, index) + `data "scalr_splunk_integration" "missing" { name = "missing-splunk-integration" }`,
				ExpectError: regexp.MustCompile("Could not find Splunk integration"),
			},
		},
	})
}

func testAccScalrSplunkIntegrationDataSourceConfig(name, endpoint, token, index string) string {
	return fmt.Sprintf(`
resource "scalr_splunk_integration" "test" {
  name     = %q
  endpoint = %q
  token    = %q
  index    = %q
  source   = "scalr-test"
}

data "scalr_splunk_integration" "by_id" {
  id = scalr_splunk_integration.test.id
}

data "scalr_splunk_integration" "by_name" {
  name = scalr_splunk_integration.test.name
}

data "scalr_splunk_integration" "by_id_and_name" {
  id   = scalr_splunk_integration.test.id
  name = scalr_splunk_integration.test.name
}
`, name, endpoint, token, index)
}
