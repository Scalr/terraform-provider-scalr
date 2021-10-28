package scalr

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccScalrIamTeam_basic(t *testing.T) {
	rInt := GetRandomInteger()
	team := &scalr.Team{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrIamTeamDestroy,
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

func TestAccScalrIamTeam_renamed(t *testing.T) {
	rInt := GetRandomInteger()
	team := &scalr.Team{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrIamTeamDestroy,
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
				PreConfig: testAccCheckScalrIamTeamRename(team),
				Config:    testAccScalrIamTeamRenamed(),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_iam_team.test", "name", "renamed-outside-of-terraform"),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrIamTeamDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrIamTeamDestroy,
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
		scalrClient := testAccProvider.Meta().(*scalr.Client)

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

func testAccCheckScalrIamTeamRename(team *scalr.Team) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		t, err := scalrClient.Teams.Read(ctx, team.ID)
		if err != nil {
			log.Fatalf("Error retrieving team: %v", err)
		}

		t, err = scalrClient.Teams.Update(
			context.Background(),
			team.ID,
			scalr.TeamUpdateOptions{
				Name:  scalr.String("renamed-outside-of-terraform"),
				Users: t.Users,
			},
		)
		if err != nil {
			log.Fatalf("Could not rename team outside of terraform: %v", err)
		}
		if t.Name != "renamed-outside-of-terraform" {
			log.Fatalf("Failed to rename the team outside of terraform: %v", err)
		}
	}
}

func testAccCheckScalrIamTeamDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

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

func testAccScalrIamTeamRenamed() string {
	return fmt.Sprintf(`
resource "scalr_iam_team" "test" {
  name        = "renamed-outside-of-terraform"
  description = "Test team"
  account_id  = "%s"
  users       = ["%s"]
}`, defaultAccount, testUser)
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
