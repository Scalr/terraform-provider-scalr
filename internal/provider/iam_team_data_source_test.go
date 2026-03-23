package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrIamTeamDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-team")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			Steps: []resource.TestStep{
				{
					Config:      `data scalr_iam_team test {}`,
					ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
					PlanOnly:    true,
				},
				{
					Config:      `data scalr_iam_team test {id = ""}`,
					ExpectError: regexp.MustCompile("Attribute id must not be empty"),
					PlanOnly:    true,
				},
				{
					Config:      `data scalr_iam_team test {name = ""}`,
					ExpectError: regexp.MustCompile("Attribute name must not be empty"),
					PlanOnly:    true,
				},
				{
					Config: testAccScalrIamTeamDataSourceByIDConfig(name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_iam_team.test", "id"),
						resource.TestCheckResourceAttr("data.scalr_iam_team.test", "name", name),
						resource.TestCheckResourceAttr(
							"data.scalr_iam_team.test",
							"description",
							"Test team description",
						),
						resource.TestCheckResourceAttr("data.scalr_iam_team.test", "account_id", defaultAccount),
						resource.TestCheckResourceAttrSet("data.scalr_iam_team.test", "identity_provider_id"),
						resource.TestCheckResourceAttr("data.scalr_iam_team.test", "users.#", "1"),
						resource.TestCheckResourceAttr("data.scalr_iam_team.test", "users.0", testUser),
					),
				},
				{
					Config: testAccScalrIamTeamDataSourceByNameConfig(name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_iam_team.test", "id"),
						resource.TestCheckResourceAttr("data.scalr_iam_team.test", "name", name),
					),
				},
				{
					Config: testAccScalrIamTeamDataSourceByIDAndNameConfig(name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_iam_team.test", "id"),
						resource.TestCheckResourceAttr("data.scalr_iam_team.test", "name", name),
					),
				},
			},
		},
	)
}

func testAccScalrIamTeamDataSourceByIDConfig(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_iam_team" "test" {
  name        = "%s"
  description = "Test team description"
  account_id  = "%s"
  users       = ["%s"]
}

data "scalr_iam_team" "test" {
  id         = scalr_iam_team.test.id
  account_id = scalr_iam_team.test.account_id
}`, name, defaultAccount, testUser,
	)
}

func testAccScalrIamTeamDataSourceByNameConfig(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_iam_team" "test" {
  name        = "%s"
  description = "Test team description"
  account_id  = "%s"
  users       = ["%s"]
}

data "scalr_iam_team" "test" {
  name       = scalr_iam_team.test.name
  account_id = scalr_iam_team.test.account_id
}`, name, defaultAccount, testUser,
	)
}

func testAccScalrIamTeamDataSourceByIDAndNameConfig(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_iam_team" "test" {
  name        = "%s"
  description = "Test team description"
  account_id  = "%s"
  users       = ["%s"]
}

data "scalr_iam_team" "test" {
  id         = scalr_iam_team.test.id
  name       = scalr_iam_team.test.name
  account_id = scalr_iam_team.test.account_id
}`, name, defaultAccount, testUser,
	)
}
