package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccWebhook_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "name", fmt.Sprintf("webhook-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "enabled", "false"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_webhook.test", "endpoint_id"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_webhook.test", "workspace_id"),
				),
			},
		},
	})
}

func TestAccWebhook_update(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "name", fmt.Sprintf("webhook-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "enabled", "false"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_webhook.test", "endpoint_id"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_webhook.test", "workspace_id"),
				),
			},
			{
				Config: testAccWebhookConfigUpdate(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "name", fmt.Sprintf("webhook-test-%d-renamed", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_webhook.test", "endpoint_id"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_webhook.test", "workspace_id"),
				),
			},
			{
				Config:      testAccWebhookConfigUpdateEmptyEvent(rInt),
				ExpectError: regexp.MustCompile("Got error during parsing events: 0-th value is empty"),
			},
		},
	})
}

func testAccWebhookConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_endpoint test {
  name         = "test endpoint-%[1]d"
  timeout      = 15
  max_attempts = 3
  url          = "https://example.com/webhook"
  environment_id = scalr_environment.test.id
}

resource scalr_webhook test {
  enabled               = false
  name                  = "webhook-test-%[1]d"
  events                = ["run:completed", "run:errored"]
  endpoint_id           = scalr_endpoint.test.id
  workspace_id          = scalr_workspace.test.id
}

data scalr_webhook test {
  id         = scalr_webhook.test.id
}`, rInt, defaultAccount)
}

func testAccWebhookConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_endpoint test {
  name         = "test endpoint-%[1]d"
  timeout      = 15
  max_attempts = 3
  url          = "https://example.com/webhook"
  environment_id = scalr_environment.test.id
}

resource scalr_webhook test {
  enabled               = true
  name                  = "webhook-test-%[1]d-renamed"
  events                = ["run:completed", "run:errored"]
  endpoint_id           = scalr_endpoint.test.id
  workspace_id          = scalr_workspace.test.id
}

data scalr_webhook test {
  id         = scalr_webhook.test.id
}`, rInt, defaultAccount)
}

func testAccWebhookConfigUpdateEmptyEvent(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_endpoint test {
  name         = "test endpoint-%[1]d"
  timeout      = 15
  max_attempts = 3
  url          = "https://example.com/webhook"
  environment_id = scalr_environment.test.id
}

resource scalr_webhook test {
  enabled               = true
  name                  = "webhook-test-%[1]d-renamed"
  events                = [""]
  endpoint_id           = scalr_endpoint.test.id
  workspace_id          = scalr_workspace.test.id
}

data scalr_webhook test {
  id         = scalr_webhook.test.id
}`, rInt, defaultAccount)
}
