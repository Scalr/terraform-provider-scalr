package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSlackIntegration_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			scalrClient, _ := createScalrClient()
			slackConnection, err := scalrClient.SlackIntegrations.GetConnection(ctx, defaultAccount)
			if err != nil {
				t.Fatalf("Error fetching Slack connection: %v", err)
				return
			}
			if slackConnection.ID == "" {
				t.Skip("Scalr instance doesn't have working slack connection.")
			}
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrSlackIntegrationConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_slack_integration.test", "id"),
					resource.TestCheckResourceAttr(
						"scalr_slack_integration.test",
						"name",
						"test-create",
					),
					resource.TestCheckResourceAttr(
						"scalr_slack_integration.test",
						"channel_id",
						"C123",
					),
					resource.TestCheckResourceAttr(
						"scalr_slack_integration.test",
						"account_id",
						defaultAccount,
					),
					resource.TestCheckResourceAttr(
						"scalr_slack_integration.test",
						"run_mode",
						"dry",
					),
					resource.TestCheckTypeSetElemAttr(
						"scalr_slack_integration.test",
						"events.*",
						"run_approval_required",
					),
					resource.TestCheckTypeSetElemAttr(
						"scalr_slack_integration.test",
						"events.*",
						"run_errored",
					),
				),
			},
			{
				Config: testAccScalrSlackIntegrationUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_slack_integration.test", "id"),
					resource.TestCheckResourceAttr(
						"scalr_slack_integration.test",
						"name",
						"test-create2",
					),
					resource.TestCheckResourceAttr(
						"scalr_slack_integration.test",
						"channel_id",
						"C123",
					),
					resource.TestCheckResourceAttr(
						"scalr_slack_integration.test",
						"account_id",
						defaultAccount,
					),
					resource.TestCheckResourceAttr(
						"scalr_slack_integration.test",
						"run_mode",
						"apply",
					),
					resource.TestCheckTypeSetElemAttr(
						"scalr_slack_integration.test",
						"events.*",
						"run_success",
					),
					resource.TestCheckTypeSetElemAttr(
						"scalr_slack_integration.test",
						"events.*",
						"run_errored",
					),
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
  run_mode       = "dry"
  events		 = ["run_approval_required", "run_errored"]
  channel_id	 = "C123"
  environments = [scalr_environment.test.id]
}`, defaultAccount)
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
  run_mode       = "apply"
  events		 = ["run_success", "run_errored"]
  channel_id	 = "C123"
  environments = [scalr_environment.test.id]
}`, defaultAccount)
}
