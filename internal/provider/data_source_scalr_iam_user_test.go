package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scalr/go-scalr"
)

func TestAccScalrIamUserDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_iam_user test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `email,id` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_iam_user test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_iam_user test {email = ""}`,
				ExpectError: regexp.MustCompile("expected \"email\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrIamUserDataSourceByIDConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "id", testUser),
					resource.TestCheckResourceAttr(
						"data.scalr_iam_user.test",
						"status",
						string(scalr.UserStatusActive),
					),
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "email", testUserEmail),
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "username", testUserEmail),
					resource.TestCheckResourceAttrSet("data.scalr_iam_user.test", "full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_iam_user.test", "teams.0"),
				),
			},
			{
				Config: testAccScalrIamUserDataSourceByEmailConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "id", testUser),
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "email", testUserEmail),
				),
			},
			{
				Config: testAccScalrIamUserDataSourceByIDAndEmailConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "id", testUser),
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "email", testUserEmail),
				),
			},
		},
	})
}

var testAccScalrIamUserDataSourceByIDConfig = fmt.Sprintf(`
data "scalr_iam_user" "test" {
  id = "%s"
}`, testUser)

var testAccScalrIamUserDataSourceByEmailConfig = fmt.Sprintf(`
data "scalr_iam_user" "test" {
  email = "%s"
}`, testUserEmail)

var testAccScalrIamUserDataSourceByIDAndEmailConfig = fmt.Sprintf(`
data "scalr_iam_user" "test" {
  id    = "%s"
  email = "%s"
}`, testUser, testUserEmail)
