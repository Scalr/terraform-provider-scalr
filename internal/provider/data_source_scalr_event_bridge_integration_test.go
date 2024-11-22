package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalrEventBridgeIntegrationDataSource_basic(t *testing.T) {
	AWSAccountId := os.Getenv("AWS_EVENT_BRIDGE_ACCOUNT_ID")
	region := os.Getenv("AWS_EVENT_BRIDGE_REGION")
	if len(AWSAccountId) == 0 || len(region) == 0 {
		t.Skip("Please set AWS_EVENT_BRIDGE_ACCOUNT_ID, AWS_EVENT_BRIDGE_REGION env variables to run this test.")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_event_bridge test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_event_bridge test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_event_bridge test {name = ""}`,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrEventBridgeDataSourceByIDConfig(AWSAccountId, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_event_bridge_integration.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_event_bridge_integration.test",
						"name",
						"test-create",
					),
					resource.TestCheckResourceAttr(
						"data.scalr_event_bridge_integration.test",
						"aws_account_id",
						AWSAccountId,
					),
					resource.TestCheckResourceAttr(
						"data.scalr_event_bridge_integration.test",
						"region",
						region,
					),
					resource.TestCheckResourceAttrSet(
						"data.scalr_event_bridge_integration.test",
						"event_source_name",
					),
					resource.TestCheckResourceAttrSet(
						"data.scalr_event_bridge_integration.test",
						"event_source_arn",
					),
				),
			},
			{
				Config: testAccScalrEventBridgeDataSourceByNameConfig(AWSAccountId, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_event_bridge_integration.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_event_bridge_integration.test",
						"name",
						"test-create",
					),
					resource.TestCheckResourceAttr(
						"data.scalr_event_bridge_integration.test",
						"aws_account_id",
						AWSAccountId,
					),
					resource.TestCheckResourceAttr(
						"data.scalr_event_bridge_integration.test",
						"region",
						region,
					),
					resource.TestCheckResourceAttrSet(
						"data.scalr_event_bridge_integration.test",
						"event_source_name",
					),
					resource.TestCheckResourceAttrSet(
						"data.scalr_event_bridge_integration.test",
						"event_source_arn",
					),
				),
			},
		},
	})
}

func testAccScalrEventBridgeDataSourceByIDConfig(awsAccountID, region string) string {
	return fmt.Sprintf(`
resource "scalr_event_bridge_integration" "test" {
  name           = "test-create"
  aws_account_id = "%s"
  region       = "%s"
}

data "scalr_event_bridge_integration" "test" {
  id       = scalr_event_bridge_integration.test.id
}
`, awsAccountID, region)
}

func testAccScalrEventBridgeDataSourceByNameConfig(awsAccountID, region string) string {
	return fmt.Sprintf(`
resource "scalr_event_bridge_integration" "test" {
  name           = "test-create"
  aws_account_id = "%s"
  region       = "%s"
}

data "scalr_event_bridge_integration" "test" {
  name       = scalr_event_bridge_integration.test.name
}
`, awsAccountID, region)
}
