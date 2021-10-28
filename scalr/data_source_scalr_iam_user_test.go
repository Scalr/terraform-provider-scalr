package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	scalr "github.com/scalr/go-scalr"
)

func TestAccScalrIamUserDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamUserDataSourceConfig(),
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
		},
	})
}

func testAccScalrIamUserDataSourceConfig() string {
	return fmt.Sprintf(`
data "scalr_iam_user" "test" {
	email = "%s"
}`, testUserEmail)
}
