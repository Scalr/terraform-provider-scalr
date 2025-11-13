package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccScalrRoleResource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRoleBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrRoleExists("scalr_role.test"),
					resource.TestCheckResourceAttr("scalr_role.test", "name", name),
					resource.TestCheckResourceAttr("scalr_role.test", "description", "test basic"),
					resource.TestCheckResourceAttr("scalr_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_role.test", "account_id", defaultAccount),
					resource.TestCheckTypeSetElemAttr("scalr_role.test", "permissions.*", "*:update"),
					resource.TestCheckTypeSetElemAttr("scalr_role.test", "permissions.*", "*:read"),
				),
			},
		},
	})
}

func TestAccScalrRoleResource_update(t *testing.T) {
	name := acctest.RandomWithPrefix("test-role")
	newName := acctest.RandomWithPrefix("test-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRoleBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_role.test", "name", name),
					resource.TestCheckResourceAttr("scalr_role.test", "description", "test basic"),
					resource.TestCheckResourceAttr("scalr_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_role.test", "account_id", defaultAccount),
					resource.TestCheckTypeSetElemAttr("scalr_role.test", "permissions.*", "*:read"),
					resource.TestCheckTypeSetElemAttr("scalr_role.test", "permissions.*", "*:update"),
				),
			},

			{
				Config: testAccScalrRoleUpdate(newName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrRoleExists("scalr_role.test"),
					resource.TestCheckResourceAttr("scalr_role.test", "name", newName),
					resource.TestCheckResourceAttr("scalr_role.test", "description", "updated"),
					resource.TestCheckResourceAttr("scalr_role.test", "is_system", "false"),
					resource.TestCheckResourceAttr("scalr_role.test", "account_id", defaultAccount),
					resource.TestCheckTypeSetElemAttr("scalr_role.test", "permissions.*", "*:update"),
					resource.TestCheckTypeSetElemAttr("scalr_role.test", "permissions.*", "*:delete"),
				),
			},
		},
	})
}

func TestAccScalrRoleResource_validation(t *testing.T) {
	name := acctest.RandomWithPrefix("test-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrRoleWithEmptyPermission(name),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("must not be empty"),
			},
		},
	})
}

func TestAccScalrRoleResource_import(t *testing.T) {
	name := acctest.RandomWithPrefix("test-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRoleBasic(name),
			},
			{
				ResourceName:      "scalr_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrRoleExists(resId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the role
		_, err := scalrClient.Role.GetRole(ctx, rs.Primary.ID, nil)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalrRoleDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_role" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Role.GetRole(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Role %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrRoleBasic(name string) string {
	return fmt.Sprintf(`
resource "scalr_role" "test" {
  name           = "%s"
  description    = "test basic"
  permissions    = [
	 "*:read",
	 "*:update"
  ]
}`, name)
}

func testAccScalrRoleUpdate(name string) string {
	return fmt.Sprintf(`
resource "scalr_role" "test" {
  name           = "%s"
  description    = "updated"
  permissions    = [
	 "*:update",
	 "*:delete"
  ]
}`, name)
}

func testAccScalrRoleWithEmptyPermission(name string) string {
	return fmt.Sprintf(`
resource "scalr_role" "test" {
  name           = "%s"
  description    = "updated"
  permissions    = [
	  "*:update",
	  ""
  ]
}`, name)
}
