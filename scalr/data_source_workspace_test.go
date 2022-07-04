package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalrWorkspaceDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "terraform_version", "0.12.19"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "working_directory", "terraform/test"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "environment_id"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "has_resources"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "created_by.0.username"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_init", "./scripts/pre-init.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_plan", "./scripts/pre-plan.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.post_plan", "./scripts/post-plan.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_apply", "./scripts/pre-apply.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.post_apply", "./scripts/post-apply.sh"),
				),
			},
		},
	})
}

func testAccScalrWorkspaceDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name                  = "workspace-test-%[1]d"
  environment_id 		= scalr_environment.test.id
  auto_apply            = true
  terraform_version     = "0.12.19"
  working_directory     = "terraform/test"
  hooks {
    pre_init   = "./scripts/pre-init.sh"
    pre_plan   = "./scripts/pre-plan.sh"
    post_plan  = "./scripts/post-plan.sh"
    pre_apply  = "./scripts/pre-apply.sh"
    post_apply = "./scripts/post-apply.sh"
  }
}

data scalr_workspace test {
  name           = scalr_workspace.test.name
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount)
}
