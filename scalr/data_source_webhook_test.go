package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWebhookDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookDataSourceConfig(rInt),
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
				Config: testAccWebhookDataSourceAccessByNameConfig(rInt),
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
				Config:      testAccWebhookDataSourceNotFoundAlmostTheSameNameConfig(rInt, cuttedRInt),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Endpoint with name 'test webhook-%s' not found", cuttedRInt)),
				PlanOnly:    true,
			},
			{
				Config:      testAccWebhookDataSourceNotFoundByNameConfig(),
				ExpectError: regexp.MustCompile("Webhook with name 'webhook-foo-bar-baz' not found or user unauthorized"),
				PlanOnly:    true,
			},
			{
				Config:      testAccWebhookNeitherNameNorIdSetConfig(),
				ExpectError: regexp.MustCompile("At least one argument 'id' or 'name' is required, but no definitions was found"),
				PlanOnly:    true,
			},
			{
				Config:      testAccWebhookBothNameAndIdSetConfig(),
				ExpectError: regexp.MustCompile("Attributes 'name' and 'id' can not be set at the same time"),
				PlanOnly:    true,
			},
		},
	})
}

func testAccWebhookDataSourceConfig(rInt int) string {
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
  name           = "test endpoint-%[1]d"
  timeout        = 15
  max_attempts   = 3
  url            = "https://example.com/webhook"
  environment_id = scalr_environment.test.id
}

resource scalr_webhook test {
  enabled      = false
  name         = "webhook-test-%[1]d"
  events       = ["run:completed", "run:errored"]
  endpoint_id  = scalr_endpoint.test.id
  workspace_id = scalr_workspace.test.id
}

data scalr_webhook test {
  id = scalr_webhook.test.id
}`, rInt, defaultAccount)
}

func testAccWebhookDataSourceAccessByNameConfig(rInt int) string {
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
  name           = "test endpoint-%[1]d"
  timeout        = 15
  max_attempts   = 3
  url            = "https://example.com/webhook"
  environment_id = scalr_environment.test.id
}

resource scalr_webhook test {
  enabled      = false
  name         = "webhook-test-%[1]d"
  events       = ["run:completed", "run:errored"]
  endpoint_id  = scalr_endpoint.test.id
  workspace_id = scalr_workspace.test.id
  environment_id = scalr_environment.test.id
}

data scalr_webhook test {
  name = scalr_webhook.test.name
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount)
}

func testAccWebhookDataSourceNotFoundByNameConfig() string {
	return `
data scalr_webhook test {
  name = "webhook-foo-bar-baz"
}`
}

func testAccWebhookNeitherNameNorIdSetConfig() string {
	return `data scalr_webhook test {}`
}

func testAccWebhookBothNameAndIdSetConfig() string {
	return `data scalr_webhook test {
		id = "foo"
		name = "bar"
	}`
}

func testAccWebhookDataSourceNotFoundAlmostTheSameNameConfig(rInt int, cuttedRInt string) string {
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
  name           = "test endpoint-%[1]d"
  timeout        = 15
  max_attempts   = 3
  url            = "https://example.com/webhook"
  environment_id = scalr_environment.test.id
}

resource scalr_webhook test {
  enabled      = false
  name         = "webhook-test-%[1]d"
  events       = ["run:completed", "run:errored"]
  endpoint_id  = scalr_endpoint.test.id
  workspace_id = scalr_workspace.test.id
}

data scalr_webhook test {
  name           = "test webhook-%s"
}`, rInt, defaultAccount, cuttedRInt)
}
