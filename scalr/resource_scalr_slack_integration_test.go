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
					resource.TestCheckResourceAttrSet("scalr_slack_integration.test", "channel_id"),
					resource.TestCheckResourceAttr("scalr_slack_integration.test", "name", slackChannelName),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func testAccScalrSlackIntegrationConfig() string {
	return fmt.Sprintf(`
data "scalr_slack_channel" "test" {
  name       = "%s"
  account_id = "%s"
}
resource scalr_environment test {
  name       = "test-env-slack"
  account_id = data.scalr_slack_channel.test.account_id
}
resource "scalr_slack_integration" "test" {
  name           = data.scalr_slack_channel.test.name
  account_id     = data.scalr_slack_channel.test.account_id
  events		 = ["run_approval_required", "run_errored"]
  channel_id	 = data.scalr_slack_channel.test.id
  "environments" = [scalr_environment.test.id]
}`, slackChannelName, defaultAccount)
}
