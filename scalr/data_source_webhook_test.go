package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccWebhookDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-wh", "name", fmt.Sprintf("webhook-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-wh", "enabled", "false"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_webhook.test-wh", "endpoint_id"),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-wh", "workspace_id", "existing-ws"),
				),
			},
		},
	})
}

func testAccWebhookDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_endpoint" "test-ep" {
  name         = "test endpoint-%d"
  http_method  = "POST"
  secret_key   = "my-secret-key" 
  timeout      = 15               
  max_attempts = 3                
  url          = "https://example.com/webhook"
  environment_id = "existing-env"
}

resource "scalr_webhook" "test-wh" {
  enabled               = false
  name                  = "webhook-test-%d"
  events                = ["run:completed", "run:errored"]
  endpoint_id           = "${scalr_endpoint.test-ep.id}"
  workspace_id          = "existing-ws"
}

data "scalr_webhook" "test-wh" {
  id         = "${scalr_webhook.test-wh.id}"
}`, rInt, rInt)
}
