package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSlackIntegration_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testSlackChannelNamePreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrSlackIntegrationConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_slack_integration.test", "id"),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "name", "test-create"),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "channel_id", slackChannelId),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "events.0", "run_approval_required"),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "events.1", "run_errored"),
				),
			},
			{
				Config: testAccScalrSlackIntegrationUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_slack_integration.test", "id"),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "name", "test-create2"),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "channel_id", slackChannelId),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "events.0", "run_success"),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "events.1", "run_errored"),
				),
			},
		},
	})
}

func testAccScalrSlackIntegrationConfig() string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-slack"
  account_id = "%s"
}
resource "scalr_slack_integration" "test" {
  name           = "test-create"
  account_id     = scalr_environment.test.account_id
  events		 = ["run_approval_required", "run_errored"]
  channel_id	 = "%s"
  environments = [scalr_environment.test.id]
}`, defaultAccount, slackChannelId)
}
func testAccScalrSlackIntegrationUpdateConfig() string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-slack"
  account_id = "%s"
}
resource "scalr_slack_integration" "test" {
  name           = "test-create2"
  account_id     = scalr_environment.test.account_id
  events		 = ["run_success", "run_errored"]
  channel_id	 = "%s"
  environments = [scalr_environment.test.id]
}`, defaultAccount, slackChannelId)
}
