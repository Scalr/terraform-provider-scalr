package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTFETeamAccessDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamAccessDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_team_access.foobar", "access", "write"),
					resource.TestCheckResourceAttrSet("data.scalr_team_access.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.scalr_team_access.foobar", "team_id"),
					resource.TestCheckResourceAttrSet("data.scalr_team_access.foobar", "workspace_id"),
				),
			},
		},
	})
}

func testAccTFETeamAccessDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "scalr_team" "foobar" {
  name         = "team-test-%d"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_workspace" "foobar" {
  name         = "workspace-test-%d"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_team_access" "foobar" {
  access       = "write"
  team_id      = "${scalr_team.foobar.id}"
  workspace_id = "${scalr_workspace.foobar.id}"
}

data "scalr_team_access" "foobar" {
  team_id      = "${scalr_team.foobar.id}"
  workspace_id = "${scalr_team_access.foobar.workspace_id}"
}`, rInt, rInt, rInt)
}
