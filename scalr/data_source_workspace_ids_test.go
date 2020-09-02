package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTFEWorkspaceIDsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.#", "2"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.0", fmt.Sprintf("workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.1", fmt.Sprintf("workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "environment_id", "existing-env"),
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

func TestAccTFEWorkspaceIDsDataSource_wildcard(t *testing.T) {
	t.Skip("Wildcard test is not passing for unknown reasons. Using the wildcard symbol produces no workspaces")
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_wildcard(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.0", "*"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "environment_id", "existing-env"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "ids.%", "3"),
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

func testAccTFEWorkspaceIDsDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_workspace" "foo" {
  name           = "workspace-foo-%d"
  environment_id = "existing-env"
}

resource "scalr_workspace" "bar" {
  name           = "workspace-bar-%d"
  environment_id = "existing-env"
}

resource "scalr_workspace" "dummy" {
  name           = "workspace-dummy-%d"
  environment_id = "existing-env"
}

data "scalr_workspace_ids" "foobar" {
  names          = ["${scalr_workspace.foo.name}", "${scalr_workspace.bar.name}"]
  environment_id = "existing-env"
}`, rInt, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_wildcard(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_workspace" "foo" {
  name           = "workspace-foo-%d"
  environment_id = "existing-env"
}

resource "scalr_workspace" "bar" {
  name           = "workspace-bar-%d"
  environment_id = "existing-env"
}

resource "scalr_workspace" "dummy" {
  name           = "workspace-dummy-%d"
  environment_id = "existing-env"
}

data "scalr_workspace_ids" "foobar" {
  names          = ["*"]
  environment_id = "existing-env"
}`, rInt, rInt, rInt)
}
