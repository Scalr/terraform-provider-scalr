package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/scalr/go-scalr/v2/scalr/schemas"
)

func TestAccScalrIamTeamResource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-team")
	team := &schemas.Team{}

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrIamTeamResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrIamTeamResourceBasic(name),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrIamTeamResourceExists("scalr_iam_team.test", team),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "name", name),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "description", "Test team"),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "account_id", defaultAccount),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "users.0", testUser),
					),
				},
			},
		},
	)
}

func TestAccScalrIamTeamResource_update(t *testing.T) {
	name := acctest.RandomWithPrefix("test-team")
	newName := acctest.RandomWithPrefix("test-team")
	team := &schemas.Team{}

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrIamTeamResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrIamTeamResourceBasic(name),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrIamTeamResourceExists("scalr_iam_team.test", team),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "name", name),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "description", "Test team"),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "account_id", defaultAccount),
						resource.TestCheckTypeSetElemAttr("scalr_iam_team.test", "users.*", testUser),
					),
				},
				{
					Config: testAccScalrIamTeamResourceUpdate(newName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrIamTeamResourceExists("scalr_iam_team.test", team),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "name", newName),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "description", "updated"),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "account_id", defaultAccount),
						resource.TestCheckResourceAttr("scalr_iam_team.test", "users.#", "0"),
					),
				},
			},
		},
	)
}

func TestAccScalrIamTeamResource_validation(t *testing.T) {
	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			Steps: []resource.TestStep{
				{
					Config:      testAccScalrIamTeamResourceEmptyUser(),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("must not be empty"),
				},
			},
		},
	)
}

func TestAccScalrIamTeamResource_import(t *testing.T) {
	name := acctest.RandomWithPrefix("test-team")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrIamTeamResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrIamTeamResourceBasic(name),
				},
				{
					ResourceName:      "scalr_iam_team.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		},
	)
}

func TestAccScalrIamTeamResource_UpgradeFromSDK(t *testing.T) {
	name := acctest.RandomWithPrefix("test-team")

	resource.Test(
		t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					ExternalProviders: map[string]resource.ExternalProvider{
						"scalr": {
							Source:            "registry.scalr.io/scalr/scalr",
							VersionConstraint: "<=3.15.0",
						},
					},
					Config: testAccScalrIamTeamResourceBasic(name),
					Check:  resource.TestCheckResourceAttrSet("scalr_iam_team.test", "id"),
				},
				{
					ProtoV5ProviderFactories: protoV5ProviderFactories(t),
					Config:                   testAccScalrIamTeamResourceBasic(name),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectEmptyPlan(),
						},
					},
				},
			},
		},
	)
}

func testAccCheckScalrIamTeamResourceExists(resId string, team *schemas.Team) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		t, err := scalrClient.Team.GetTeam(ctx, rs.Primary.ID, nil)
		if err != nil {
			return err
		}

		*team = *t

		return nil
	}
}

func testAccCheckScalrIamTeamResourceDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_iam_team" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Team.GetTeam(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Team %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrIamTeamResourceBasic(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_iam_team" "test" {
  name        = "%s"
  description = "Test team"
  account_id  = "%s"
  users       = ["%s"]
}`, name, defaultAccount, testUser,
	)
}

func testAccScalrIamTeamResourceUpdate(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_iam_team" "test" {
  name        = "%s"
  description = "updated"
  account_id  = "%s"
  users       = []
}`, name, defaultAccount,
	)
}

func testAccScalrIamTeamResourceEmptyUser() string {
	return fmt.Sprintf(
		`
resource "scalr_iam_team" "test" {
  name        = "test-team"
  account_id  = "%s"
  users       = ["%s", ""]
}`, defaultAccount, testUser,
	)
}
