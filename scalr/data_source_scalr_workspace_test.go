package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalrWorkspaceDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrWorkspaceDataSourceMissingRequiredConfig,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      testAccScalrWorkspaceDataSourceIDIsEmptyConfig,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      testAccScalrWorkspaceDataSourceNameIsEmptyConfig,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrWorkspaceDataSourceByIDConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "terraform_version", "1.1.9"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "iac_platform", "terraform"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "working_directory", "terraform/test"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "auto_queue_runs", "skip_first"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "deletion_protection_enabled", "true"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "environment_id"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "has_resources"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "created_by.0.username"),
					resource.TestCheckResourceAttr("data.scalr_workspace.test", "tags.#", "0"),
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
			{
				Config: testAccScalrWorkspaceDataSourceByNameConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "environment_id"),
				),
			},
			{
				Config: testAccScalrWorkspaceDataSourceByIDAndNameConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.test", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.test", "environment_id"),
				),
			},
		},
	})
}

var testAccScalrWorkspaceDataSourceMissingRequiredConfig = `
data scalr_workspace test {
  environment_id = "test-env-id"
}`

var testAccScalrWorkspaceDataSourceIDIsEmptyConfig = `
data scalr_workspace test {
  id             = ""
  environment_id = "test-env-id"
}`

var testAccScalrWorkspaceDataSourceNameIsEmptyConfig = `
data scalr_workspace test {
  name           = ""
  environment_id = "test-env-id"
}`

func testAccScalrWorkspaceDataSourceByIDConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name                  = "workspace-test-%[1]d"
  environment_id 		= scalr_environment.test.id
  auto_apply            = true
  terraform_version     = "1.1.9"
  iac_platform          = "terraform"
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
  id             = scalr_workspace.test.id
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount)
}

func testAccScalrWorkspaceDataSourceByNameConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name                  = "workspace-test-%[1]d"
  environment_id 		= scalr_environment.test.id
}

data scalr_workspace test {
  name           = scalr_workspace.test.name
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount)
}

func testAccScalrWorkspaceDataSourceByIDAndNameConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name                  = "workspace-test-%[1]d"
  environment_id 		= scalr_environment.test.id
}

data scalr_workspace test {
  id             = scalr_workspace.test.id
  name           = scalr_workspace.test.name
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount)
}
