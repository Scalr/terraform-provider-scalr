package scalr

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccScalrIamRole_basic(t *testing.T) {
	role := &scalr.Role{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrIamRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamRoleBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrIamRoleExists("scalr_iam_role.test", role),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "name", "role-test"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "description", "test basic"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.1", "*:update"),
				),
			},
		},
	})
}

func TestAccScalrIamRole_renamed(t *testing.T) {
	role := &scalr.Role{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrIamRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamRoleBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrIamRoleExists("scalr_iam_role.test", role),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "name", "role-test"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "description", "test basic"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.1", "*:update"),
				),
			},

			{
				PreConfig: testAccCheckScalrIamRoleRename(role),
				Config:    testAccScalrIamRoleRenamed(),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_iam_role.test", "name", "renamed-outside-of-terraform"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "description", "test basic"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.1", "*:update"),
				),
			},
		},
	})
}
func TestAccScalrIamRole_update(t *testing.T) {
	role := &scalr.Role{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrIamRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamRoleBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_iam_role.test", "name", "role-test"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "description", "test basic"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.0", "*:read"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.1", "*:update"),
				),
			},

			{
				Config: testAccScalrIamRoleUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrIamRoleExists("scalr_iam_role.test", role),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "name", "role-updated"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "description", "updated"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.0", "*:update"),
					resource.TestCheckResourceAttr("scalr_iam_role.test", "permissions.1", "*:delete"),
				),
			},
		},
	})
}

func TestAccScalrIamRole_import(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrIamRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamRoleBasic(),
			},

			{
				ResourceName:      "scalr_iam_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrIamRoleExists(resId string, role *scalr.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

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

func testAccCheckScalrIamRoleRename(role *scalr.Role) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		r, err := scalrClient.Roles.Read(ctx, role.ID)

		if err != nil {
			log.Fatalf("Error retrieving role: %v", err)
		}

		r, err = scalrClient.Roles.Update(
			context.Background(),
			r.ID,
			scalr.RoleUpdateOptions{Name: scalr.String("renamed-outside-of-terraform")},
		)
		if err != nil {
			log.Fatalf("Could not rename the role outside of terraform: %v", err)
		}

		if r.Name != "renamed-outside-of-terraform" {
			log.Fatalf("Failed to rename the role outside of terraform: %v", err)
		}
	}
}

func testAccCheckScalrIamRoleDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_iam_role" {
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

func testAccScalrIamRoleBasic() string {
	return fmt.Sprintf(`
resource "scalr_iam_role" "test" {
  name           = "role-test"
  description    = "test basic"
  account_id     = "%s"
  permissions    = [
	 "*:read",
	 "*:update"
  ]
}`, defaultAccount)
}

func testAccScalrIamRoleRenamed() string {
	return fmt.Sprintf(`
resource "scalr_iam_role" "test" {
  name           = "renamed-outside-of-terraform"
  description    = "test basic"
  account_id     = "%s"
  permissions    = [
	 "*:read",
	 "*:update"
  ]
}`, defaultAccount)
}

func testAccScalrIamRoleUpdate() string {
	return fmt.Sprintf(`
resource "scalr_iam_role" "test" {
  name           = "role-updated"
  account_id     = "%s"
  description    = "updated"
  permissions    = [
	 "*:update",
	 "*:delete"
  ]
}`, defaultAccount)
}
