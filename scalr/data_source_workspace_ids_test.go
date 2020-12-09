package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalrWorkspaceIDsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceIDsDataSourceConfigBasic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.#", "2"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.0", fmt.Sprintf("workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.1", fmt.Sprintf("workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", "environment_id"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "ids.%", "2"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_workspace_ids.foobar", "id"),
				),
			},
		},
	})
}

func TestAccScalrWorkspaceIDsDataSource_wildcard(t *testing.T) {
	t.Skip("Wildcard test is not passing for unknown reasons. Using the wildcard symbol produces no workspaces")
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceIDsDataSourceConfigWildcard(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.0", "*"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "ids.%", "3"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace_ids.foobar", "environment_id"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-dummy-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_workspace_ids.foobar", "id"),
				),
			},
		},
	})
}

func testAccScalrWorkspaceIDsDataSourceConfigBasic(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace foo {
  name           = "workspace-foo-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace bar {
  name           = "workspace-bar-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace dummy {
  name           = "workspace-dummy-%[1]d"
  environment_id = scalr_environment.test.id
}

data scalr_workspace_ids foobar {
  names          = [scalr_workspace.foo.name, scalr_workspace.bar.name]
  environment_id = scalr_environment.test.id
}`, rInt, DefaultAccount)
}

func testAccScalrWorkspaceIDsDataSourceConfigWildcard(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace foo {
  name           = "workspace-foo-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace bar {
  name           = "workspace-bar-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace dummy {
  name           = "workspace-dummy-%[1]d"
  environment_id = scalr_environment.test.id
}

data scalr_workspace_ids foobar {
  names          = ["*"]
  environment_id = scalr_environment.test.id
}`, rInt, DefaultAccount)
}
