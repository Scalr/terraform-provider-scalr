package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrIamTeamDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_iam_team test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_iam_team test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_iam_team test {name = ""}`,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrIamTeamDataSourceByIDConfig(rInt),
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
			{
				Config: testAccScalrIamTeamDataSourceByNameConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_iam_team.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_iam_team.test",
						"name",
						fmt.Sprintf("test-team-%d", rInt),
					),
				),
			},
			{
				Config: testAccScalrIamTeamDataSourceByIDAndNameConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_iam_team.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_iam_team.test",
						"name",
						fmt.Sprintf("test-team-%d", rInt),
					),
				),
			},
		},
	})
}

func testAccScalrIamTeamDataSourceByIDConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_iam_team" "test" {
  name       = "test-team-%d"
  account_id = "%s"
  users      = ["%s"]
}

data "scalr_iam_team" "test" {
  id         = scalr_iam_team.test.id
  account_id = scalr_iam_team.test.account_id
}`, rInt, defaultAccount, testUser)
}

func testAccScalrIamTeamDataSourceByNameConfig(rInt int) string {
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

func testAccScalrIamTeamDataSourceByIDAndNameConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_iam_team" "test" {
  name       = "test-team-%d"
  account_id = "%s"
  users      = ["%s"]
}

data "scalr_iam_team" "test" {
  id         = scalr_iam_team.test.id
  name       = scalr_iam_team.test.name
  account_id = scalr_iam_team.test.account_id
}`, rInt, defaultAccount, testUser)
}
