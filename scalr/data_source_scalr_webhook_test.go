package scalr

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccWebhookDataSource_basic(t *testing.T) {
	rInt := rand.Intn(100)

	cutRInt := strconv.Itoa(rInt)[:len(strconv.Itoa(rInt))-1]

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_webhook test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_webhook test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_webhook test {name = ""}`,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccWebhookDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-new", "name", fmt.Sprintf("webhook-test-new-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-new", "enabled", "false"),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-new", "url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-new", "secret_key", ""),
					resource.TestCheckResourceAttrSet(
						"data.scalr_webhook.test-new", "timeout"),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-new", "max_attempts", "2"),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-new", "header.#", "2"),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test-new", "environments.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_webhook.test-new",
						"environments.0",
						"scalr_environment.test-new",
						"id"),
				),
			},
			{
				Config: testAccWebhookDataSourceAccessByNameConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "name", fmt.Sprintf("webhook-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "enabled", "false"),
				),
			},
			{
				Config: testAccWebhookDataSourceAccessByIDAndNameConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "name", fmt.Sprintf("webhook-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_webhook.test", "enabled", "false"),
				),
			},
			{
				Config:      testAccWebhookDataSourceNotFoundAlmostTheSameNameConfig(rInt, cutRInt),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Webhook with name 'test webhook-%s' not found", cutRInt)),
				PlanOnly:    true,
			},
			{
				Config:      testAccWebhookDataSourceNotFoundByNameConfig(),
				ExpectError: regexp.MustCompile("Webhook with name 'webhook-foo-bar-baz' not found or user unauthorized"),
				PlanOnly:    true,
			},
		},
	})
}

func testAccWebhookDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test-new {
  name       = "test-env-new-%[1]d"
  account_id = "%s"
}

resource scalr_webhook test-new {
  account_id   = "%[2]s"
  enabled      = false
  name         = "webhook-test-new-%[1]d"
  events       = ["run:completed", "run:errored"]
  environments = [scalr_environment.test-new.id]
  url          = "https://example.com/webhook"
  max_attempts = 2
  header {
    name  = "header-1"
    value = "value-1"
  }
  header {
    name  = "header-2"
    value = "value-2"
  }
}

data scalr_webhook test-new {
  id = scalr_webhook.test-new.id
}`, rInt, defaultAccount)
}

func testAccWebhookDataSourceAccessByNameConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_webhook test {
  enabled      = false
  name         = "webhook-test-%[1]d"
  events       = ["run:completed", "run:errored"]
  url          = "https://example.com/webhook"
  account_id   = "%s"
}

data scalr_webhook test {
  name       = scalr_webhook.test.name
  account_id = "%s"
}`, rInt, defaultAccount, defaultAccount)
}

func testAccWebhookDataSourceAccessByIDAndNameConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_webhook test {
  enabled      = false
  name         = "webhook-test-%[1]d"
  events       = ["run:completed", "run:errored"]
  url          = "https://example.com/webhook"
  account_id   = "%s"
}

data scalr_webhook test {
  id         = scalr_webhook.test.id
  name       = scalr_webhook.test.name
  account_id = "%s"
}`, rInt, defaultAccount, defaultAccount)
}

func testAccWebhookDataSourceNotFoundByNameConfig() string {
	return `
data scalr_webhook test {
  name       = "webhook-foo-bar-baz"
  account_id = "foobar"
}`
}

func testAccWebhookDataSourceNotFoundAlmostTheSameNameConfig(rInt int, cutRInt string) string {
	return fmt.Sprintf(`
resource scalr_webhook test {
  enabled      = false
  name         = "test webhook-%[1]d"
  events       = ["run:completed", "run:errored"]
  url          = "https://example.com/webhook"	
  account_id   = "%s"
}

data scalr_webhook test {
  name       = "test webhook-%[3]s"
  account_id = "%s"
}`, rInt, defaultAccount, cutRInt, defaultAccount)
}
