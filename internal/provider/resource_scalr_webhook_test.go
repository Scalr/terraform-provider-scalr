package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWebhook_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "name", fmt.Sprintf("webhook-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccWebhook_update(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccWebhookConfigUpdateEmptyEvent(rInt),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("expected events to be one of"),
			},
			{
				Config: testAccWebhookConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "name", fmt.Sprintf("webhook-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "enabled", "false"),
				),
			},
			{
				Config: testAccWebhookConfigUpdate(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "name", fmt.Sprintf("webhook-test-%d-renamed", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "enabled", "true"),
				),
			},
		},
	})
}

func testAccWebhookConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_webhook test {
  enabled               = false
  name                  = "webhook-test-%[1]d"
  events                = ["run:completed", "run:errored"]
  url                   = "https://example.com/webhook"
  account_id            = "%s"
}

data scalr_webhook test {
  id         = scalr_webhook.test.id
}`, rInt, defaultAccount)
}

func testAccWebhookConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource scalr_webhook test {
  enabled               = true
  name                  = "webhook-test-%[1]d-renamed"
  events                = ["run:completed", "run:errored"]
  url                   = "https://example.com/webhook"
  account_id            = "%s"
}

data scalr_webhook test {
  id         = scalr_webhook.test.id
}`, rInt, defaultAccount)
}

func testAccWebhookConfigUpdateEmptyEvent(rInt int) string {
	return fmt.Sprintf(`
resource scalr_webhook test {
  enabled               = true
  name                  = "webhook-test-%[1]d-renamed"
  events                = [""]
  account_id            = "%s"
  url                   = "https://example.com/webhook"
}

data scalr_webhook test {
  id         = scalr_webhook.test.id
}`, rInt, defaultAccount)
}
