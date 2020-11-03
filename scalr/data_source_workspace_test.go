package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalrWorkspaceDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_workspace.foobar", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.foobar", "queue_all_runs", "false"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.foobar", "terraform_version", "0.12.19"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace.foobar", "working_directory", "terraform/test"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.foobar", "environment_id"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.foobar", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.foobar", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace.foobar", "created_by.0.username"),
				),
			},
		},
	})
}

func testAccScalrWorkspaceDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env"
  account_id = "acc-svrcncgh453bi8g"
}

resource scalr_workspace foobar {
  name                  = "workspace-test-%d"
  environment_id 		= scalr_environment.test.id
  auto_apply            = true
  queue_all_runs        = false
  terraform_version     = "0.12.19"
  working_directory     = "terraform/test"
}

data scalr_workspace foobar {
  name           = scalr_workspace.foobar.name
  environment_id = scalr_environment.test.id
}`, rInt)
}
