package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEventBridgeIntegration_basic(t *testing.T) {
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
				Config: testAccScalrEventBridgeIntegrationConfig(AWSAccountId, region),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_event_bridge_integration.test", "id"),
					resource.TestCheckResourceAttr(
						"scalr_event_bridge_integration.test",
						"name",
						"test-create",
					),
					resource.TestCheckResourceAttr(
						"scalr_event_bridge_integration.test",
						"aws_account_id",
						AWSAccountId,
					),
					resource.TestCheckResourceAttr(
						"scalr_event_bridge_integration.test",
						"region",
						region,
					),
					resource.TestCheckResourceAttrSet(
						"scalr_event_bridge_integration.test",
						"event_source_name",
					),
					resource.TestCheckResourceAttrSet(
						"scalr_event_bridge_integration.test",
						"event_source_arn",
					),
				),
			},
		},
	})
}

func testAccScalrEventBridgeIntegrationConfig(awsAccountID, region string) string {
	return fmt.Sprintf(`
resource "scalr_event_bridge_integration" "test" {
  name           = "test-create"
  aws_account_id = "%s"
  region       = "%s"
}`, awsAccountID, region)
}
