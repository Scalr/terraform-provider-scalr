package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalrIamRoleDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamRoleDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_iam_role.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_iam_role.test", "name", "role-test"),
					resource.TestCheckResourceAttr("data.scalr_iam_role.test", "description", ""),
					resource.TestCheckResourceAttr("data.scalr_iam_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("data.scalr_iam_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("data.scalr_iam_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("data.scalr_iam_role.test", "permissions.1", "roles:update"),
				),
			},
		},
	})
}

func testAccScalrIamRoleDataSourceConfig() string {
	return fmt.Sprintf(`
resource "scalr_iam_role" "test" {
  name             = "role-test"
  account_id       = "%s"
  permissions      = [
    "*:read",
	"roles:update"
  ]
}

data "scalr_iam_role" "test" {
	id = scalr_iam_role.test.id
}`, defaultAccount)
}
