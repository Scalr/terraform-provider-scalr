package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalrIamTeamDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamTeamDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_iam_team.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_iam_team.test",
						"name",
						fmt.Sprintf("test-team-%d", rInt),
					),
					resource.TestCheckResourceAttr("data.scalr_iam_team.test", "description", ""),
					resource.TestCheckResourceAttr("data.scalr_iam_team.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttrSet("data.scalr_iam_team.test", "identity_provider_id"),
					resource.TestCheckResourceAttr("data.scalr_iam_team.test", "users.0", testUser),
				),
			},
		},
	})
}

func testAccScalrIamTeamDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_iam_team" "test" {
  name       = "test-team-%d"
  account_id = "%s"
  users      = ["%s"]
}

data "scalr_iam_team" "test" {
	name       = scalr_iam_team.test.name
	account_id = scalr_iam_team.test.account_id
}`, rInt, defaultAccount, testUser)
}
