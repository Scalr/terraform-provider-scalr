package scalr

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	tfe "github.com/scalr/go-scalr"
)

func TestAccTFETeamMembers_basic(t *testing.T) {
	users := []*tfe.User{}
	TFE_USER1_HASH := hashSchemaString(TFE_USER1)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if TFE_USER1 == "" {
				t.Skip("Please set TFE_USER1 to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"scalr_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", TFE_USER1}),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", "usernames.3672628397", "admin"),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", fmt.Sprintf("usernames.%d", TFE_USER1_HASH), TFE_USER1),
				),
			},
		},
	})
}

func TestAccTFETeamMembers_update(t *testing.T) {
	users := []*tfe.User{}
	TFE_USER1_HASH := hashSchemaString(TFE_USER1)
	TFE_USER2_HASH := hashSchemaString(TFE_USER2)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if TFE_USER1 == "" {
				t.Skip("Please set TFE_USER1 to run this test")
			}
			if TFE_USER2 == "" {
				t.Skip("Please set TFE_USER2 to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"scalr_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", TFE_USER1}),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", "usernames.3672628397", "admin"),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", fmt.Sprintf("usernames.%d", TFE_USER1_HASH), TFE_USER1),
				),
			},

			{
				Config: testAccTFETeamMembers_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFETeamMembersExists(
						"scalr_team_members.foobar", &users),
					testAccCheckTFETeamMembersAttributes(&users, []string{"admin", TFE_USER2}),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", "usernames.#", "2"),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", fmt.Sprintf("usernames.%d", TFE_USER2_HASH), TFE_USER2),
					resource.TestCheckResourceAttr(
						"scalr_team_members.foobar", "usernames.3672628397", "admin"),
				),
			},
		},
	})
}

func TestAccTFETeamMembers_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if TFE_USER1 == "" {
				t.Skip("Please set TFE_USER1 to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamMembersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETeamMembers_basic,
			},

			{
				ResourceName:      "scalr_team_members.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hashSchemaString(username string) int {
	return schema.HashSchema(&schema.Schema{Type: schema.TypeString})(username)

}

func testAccCheckTFETeamMembersExists(
	n string, users *[]*tfe.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		us, err := tfeClient.TeamMembers.List(ctx, rs.Primary.ID)
		if err != nil && err != tfe.ErrResourceNotFound {
			return err
		}

		if len(us) != 2 {
			return fmt.Errorf("Users not found: %#+v", us[0])
		}

		*users = us

		return nil
	}
}

func testAccCheckTFETeamMembersAttributes(
	users *[]*tfe.User, expectedUsernames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usernames := usernamesFromTFEUsers(*users)
		if !reflect.DeepEqual(usernames, expectedUsernames) {
			return fmt.Errorf("Expected usernames: %q, Given: %q",
				expectedUsernames, usernames)
		}

		return nil
	}
}

func usernamesFromTFEUsers(users []*tfe.User) []string {
	usernames := make([]string, len(users), len(users))
	for i, user := range users {
		usernames[i] = user.Username
	}
	return usernames
}

func testAccCheckTFETeamMembersDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_team_members" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		users, err := tfeClient.TeamMembers.List(ctx, rs.Primary.ID)
		if err != nil && err != tfe.ErrResourceNotFound {
			return err
		}

		if len(users) != 0 {
			return fmt.Errorf("Users still exist")
		}
	}

	return nil
}

var testAccTFETeamMembers_basic = fmt.Sprintf(`
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_team" "foobar" {
  name         = "team-test"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_team_members" "foobar" {
  team_id   = "${scalr_team.foobar.id}"
  usernames = ["%s"]
}`, TFE_USER1)

var testAccTFETeamMembers_update = fmt.Sprintf(`
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_team" "foobar" {
  name         = "team-test"
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_team_members" "foobar" {
  team_id   = "${scalr_team.foobar.id}"
  usernames = ["%s", "%s"]
}`, TFE_USER1, TFE_USER2)
