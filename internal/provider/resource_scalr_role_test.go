package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrRole_basic(t *testing.T) {
	role := &scalr.Role{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRoleBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrRoleExists("scalr_role.test", role),
					resource.TestCheckResourceAttr("scalr_role.test", "name", "role-test"),
					resource.TestCheckResourceAttr("scalr_role.test", "description", "test basic"),
					resource.TestCheckResourceAttr("scalr_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("scalr_role.test", "permissions.1", "*:update"),
				),
			},
		},
	})
}

func TestAccScalrRole_update(t *testing.T) {
	role := &scalr.Role{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRoleBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_role.test", "name", "role-test"),
					resource.TestCheckResourceAttr("scalr_role.test", "description", "test basic"),
					resource.TestCheckResourceAttr("scalr_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("scalr_role.test", "permissions.1", "*:update"),
				),
			},

			{
				Config: testAccScalrRoleUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrRoleExists("scalr_role.test", role),
					resource.TestCheckResourceAttr("scalr_role.test", "name", "role-updated"),
					resource.TestCheckResourceAttr("scalr_role.test", "description", "updated"),
					resource.TestCheckResourceAttr("scalr_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_role.test", "permissions.0", "*:update"),
					resource.TestCheckResourceAttr("scalr_role.test", "permissions.1", "*:delete"),
				),
			},

			{
				Config:      testAccScalrRoleUpdateEmptyPermission(),
				ExpectError: regexp.MustCompile("Got error during parsing permissions: 1-th value is empty"),
			},
		},
	})
}

func TestAccScalrRole_import(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRoleBasic(),
			},

			{
				ResourceName:      "scalr_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrRoleExists(resId string, role *scalr.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the role
		r, err := scalrClient.Roles.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*role = *r

		return nil
	}
}

func testAccCheckScalrRoleDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_role" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Roles.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Role %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrRoleBasic() string {
	return `
resource "scalr_role" "test" {
  name           = "role-test"
  description    = "test basic"
  permissions    = [
	 "*:read",
	 "*:update"
  ]
}`
}

func testAccScalrRoleUpdate() string {
	return `
resource "scalr_role" "test" {
  name           = "role-updated"
  description    = "updated"
  permissions    = [
	 "*:update",
	 "*:delete"
  ]
}`
}

func testAccScalrRoleUpdateEmptyPermission() string {
	return `
resource "scalr_role" "test" {
  name           = "role-updated"
  description    = "updated"
  permissions    = [
	  "*:update",
	  ""
  ]
}`
}
