package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrIamTeam_basic(t *testing.T) {
	rInt := GetRandomInteger()
	team := &scalr.Team{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrIamTeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamTeamBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrIamTeamExists("scalr_iam_team.test", team),
					resource.TestCheckResourceAttr(
						"scalr_iam_team.test",
						"name",
						fmt.Sprintf("test-team-%d", rInt),
					),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "description", "Test team"),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "users.0", testUser),
				),
			},
		},
	})
}

func TestAccScalrIamTeam_update(t *testing.T) {
	rInt := GetRandomInteger()
	team := &scalr.Team{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrIamTeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamTeamBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrIamTeamExists("scalr_iam_team.test", team),
					resource.TestCheckResourceAttr(
						"scalr_iam_team.test",
						"name",
						fmt.Sprintf("test-team-%d", rInt),
					),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "description", "Test team"),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "users.0", testUser),
				),
			},
			{
				Config: testAccScalrIamTeamUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrIamTeamExists("scalr_iam_team.test", team),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "name", "team-updated"),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "description", "updated"),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_iam_team.test", "users.0", testUser),
				),
			},
			{
				Config:      testAccScalrIamTeamUpdateEmptyUser(),
				ExpectError: regexp.MustCompile("Got error during parsing users: 1-th value is empty"),
			},
		},
	})
}

func TestAccScalrIamTeam_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrIamTeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIamTeamBasic(rInt),
			},
			{
				ResourceName:      "scalr_iam_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrIamTeamExists(resId string, team *scalr.Team) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		t, err := scalrClient.Teams.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*team = *t

		return nil
	}
}

func testAccCheckScalrIamTeamDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_iam_team" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Teams.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Team %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrIamTeamBasic(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_iam_team" "test" {
  name        = "test-team-%d"
  description = "Test team"
  account_id  = "%s"
  users       = ["%s"]
}`, rInt, defaultAccount, testUser)
}

func testAccScalrIamTeamUpdate() string {
	return fmt.Sprintf(`
resource "scalr_iam_team" "test" {
  name        = "team-updated"
  description = "updated"
  account_id  = "%s"
  users       = ["%s"]
}`, defaultAccount, testUser)
}

func testAccScalrIamTeamUpdateEmptyUser() string {
	return fmt.Sprintf(`
resource "scalr_iam_team" "test" {
  name        = "team-updated"
  description = "updated"
  account_id  = "%s"
  users       = ["%s", ""]
}`, defaultAccount, testUser)
}
