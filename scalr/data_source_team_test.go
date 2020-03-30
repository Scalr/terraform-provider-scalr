package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTFETeamDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_team.foobar", "name", fmt.Sprintf("team-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_team.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_team.foobar", "id"),
				),
			},
		},
	})
}

func testAccTFETeamDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "scalr_team" "foobar" {
  name         = "team-test-%d"
  organization = "${scalr_organization.foobar.id}"
}

data "scalr_team" "foobar" {
  name         = "${scalr_team.foobar.name}"
  organization = "${scalr_team.foobar.organization}"
}`, rInt, rInt)
}
