package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scalr/go-scalr"

	scalr2 "github.com/scalr/terraform-provider-scalr/scalr"
)

func TestAccScalrIamUserDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { scalr2.testAccPreCheck(t) },
		ProviderFactories: scalr2.testAccProviderFactories,
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
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "id", scalr2.testUser),
					resource.TestCheckResourceAttr(
						"data.scalr_iam_user.test",
						"status",
						string(scalr.UserStatusActive),
					),
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "email", scalr2.testUserEmail),
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "username", scalr2.testUserEmail),
					resource.TestCheckResourceAttrSet("data.scalr_iam_user.test", "full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_iam_user.test", "teams.0"),
				),
			},
			{
				Config: testAccScalrIamUserDataSourceByEmailConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "id", scalr2.testUser),
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "email", scalr2.testUserEmail),
				),
			},
			{
				Config: testAccScalrIamUserDataSourceByIDAndEmailConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "id", scalr2.testUser),
					resource.TestCheckResourceAttr("data.scalr_iam_user.test", "email", scalr2.testUserEmail),
				),
			},
		},
	})
}

var testAccScalrIamUserDataSourceByIDConfig = fmt.Sprintf(`
data "scalr_iam_user" "test" {
  id = "%s"
}`, scalr2.testUser)

var testAccScalrIamUserDataSourceByEmailConfig = fmt.Sprintf(`
data "scalr_iam_user" "test" {
  email = "%s"
}`, scalr2.testUserEmail)

var testAccScalrIamUserDataSourceByIDAndEmailConfig = fmt.Sprintf(`
data "scalr_iam_user" "test" {
  id    = "%s"
  email = "%s"
}`, scalr2.testUser, scalr2.testUserEmail)
