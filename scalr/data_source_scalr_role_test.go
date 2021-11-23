package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccScalrRoleDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRoleDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_role.test", "scalr_role.test"),
					resource.TestCheckResourceAttrSet("data.scalr_role.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "name", "role-test"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "description", ""),
					resource.TestCheckResourceAttr("data.scalr_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("data.scalr_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "permissions.1", "roles:update"),

					resource.TestCheckResourceAttr("data.scalr_role.user", "id", userRole),
					resource.TestCheckResourceAttr("data.scalr_role.user", "name", "user"),
					resource.TestCheckResourceAttrSet("data.scalr_role.user", "description"),
					resource.TestCheckResourceAttr("data.scalr_role.user", "is_system", "true"),
					resource.TestCheckNoResourceAttr("data.scalr_role.user", "account_id"),
				),
			},
		},
	})
}

func testAccCheckEqualID(dataSourceId, resourceId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceId]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceId)
		}

		ds, ok := s.RootModule().Resources[dataSourceId]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceId)
		}

		if rs.Primary.ID != ds.Primary.ID {
			return fmt.Errorf("Data source returned wrong object ID: %s != %s", rs.Primary.ID, ds.Primary.ID)
		}

		return nil
	}
}

func testAccScalrRoleDataSourceConfig() string {
	return fmt.Sprintf(`
resource "scalr_role" "test" {
  name             = "role-test"
  account_id       = "%s"
  permissions      = [
    "*:read",
	"roles:update"
  ]
}

data "scalr_role" "test" {
  name       = scalr_role.test.name
  account_id = scalr_role.test.account_id
}

data "scalr_role" "user" {
    name = "user"
}
`, defaultAccount)
}
