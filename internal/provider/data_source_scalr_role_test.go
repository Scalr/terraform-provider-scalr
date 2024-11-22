package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccScalrRoleDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_role test_role {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_role test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_role test {name = ""}`,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrRoleDataSourceByIDConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_role.test", "scalr_role.test"),
					resource.TestCheckResourceAttrSet("data.scalr_role.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "name", "role-test"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "description", ""),
					resource.TestCheckResourceAttr("data.scalr_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("data.scalr_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "permissions.1", "roles:update"),
				),
			},
			{
				Config: testAccScalrRoleDataSourceByNameConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_role.test", "scalr_role.test"),
					resource.TestCheckResourceAttrSet("data.scalr_role.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "name", "role-test"),
				),
			},
			{
				Config: testAccScalrRoleDataSourceByIDAndNameConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_role.test", "scalr_role.test"),
					resource.TestCheckResourceAttrSet("data.scalr_role.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_role.test", "name", "role-test"),
				),
			},
			{
				Config: testAccScalrRoleDataSourceUserConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_role.user", "id", userRole),
					resource.TestCheckResourceAttr("data.scalr_role.user", "name", "user"),
					resource.TestCheckResourceAttrSet("data.scalr_role.user", "description"),
					resource.TestCheckResourceAttr("data.scalr_role.user", "is_system", "true"),
					resource.TestCheckResourceAttr("data.scalr_role.user", "account_id", ""),
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

var testAccScalrRoleDataSourceByIDConfig = fmt.Sprintf(`
resource "scalr_role" "test" {
  name             = "role-test"
  account_id       = "%s"
  permissions      = [
    "*:read",
	"roles:update"
  ]
}

data "scalr_role" "test" {
  id       = scalr_role.test.id
  account_id = scalr_role.test.account_id
}
`, defaultAccount)

var testAccScalrRoleDataSourceByNameConfig = fmt.Sprintf(`
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
`, defaultAccount)

var testAccScalrRoleDataSourceByIDAndNameConfig = fmt.Sprintf(`
resource "scalr_role" "test" {
  name             = "role-test"
  account_id       = "%s"
  permissions      = [
    "*:read",
	"roles:update"
  ]
}

data "scalr_role" "test" {
  id         = scalr_role.test.id
  name       = scalr_role.test.name
  account_id = scalr_role.test.account_id
}
`, defaultAccount)

func testAccScalrRoleDataSourceUserConfig() string {
	return `
data "scalr_role" "user" {
	name = "user"
}
`
}
