package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestScalrWorkspaceRunSchedule_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceRunSchedule(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_workspace_run_schedule.test", "apply_schedule", "30 3 5 3-5 2"),
					resource.TestCheckResourceAttr(
						"scalr_workspace_run_schedule.test", "destroy_schedule", "30 4 5 3-5 2"),
				),
			},
		},
	})
}

func TestScalrWorkspaceRunSchedule_default(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceRunScheduleDefaultValue(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_workspace_run_schedule.test", "apply_schedule", "0 22 * * 1-5"),
					resource.TestCheckResourceAttr(
						"scalr_workspace_run_schedule.test", "destroy_schedule", ""),
				),
			},
		},
	})
}

const testScalrWorkspaceRunScheduleCommonConfig = `
resource scalr_environment test {
  name       = "test-env-rs-%d"
  account_id = "%s"
}
resource scalr_workspace test {
  name                   = "workspace-run-schedule-test"
  environment_id         = scalr_environment.test.id
  auto_apply             = true
  run_operation_timeout = 18
  hooks {
    pre_plan   = "./scripts/pre-plan.sh"
    post_plan  = "./scripts/post-plan.sh"
    pre_apply  = "./scripts/pre-apply.sh"
    post_apply = "./scripts/post-apply.sh"
  }
}
%s
`

func testAccScalrWorkspaceRunSchedule(rInt int) string {
	return fmt.Sprintf(testScalrWorkspaceRunScheduleCommonConfig, rInt, defaultAccount, `
resource scalr_workspace_run_schedule test {
	workspace_id = scalr_workspace.test.id
	apply_schedule = "30 3 5 3-5 2"
	destroy_schedule = "30 4 5 3-5 2"
}`)
}

func testAccScalrWorkspaceRunScheduleDefaultValue(rInt int) string {
	return fmt.Sprintf(testScalrWorkspaceRunScheduleCommonConfig, rInt, defaultAccount, `
resource scalr_workspace_run_schedule test {
	workspace_id = scalr_workspace.test.id
	apply_schedule = "0 22 * * 1-5"
}`)
}
