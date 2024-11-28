package provider

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

const (
	policyGroupVcsRepoID   = "Scalr/tf-revizor-fixtures"
	policyGroupVcsRepoPath = "policies/clouds"
)

func TestAccPolicyGroup_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// TODO: delete skip after SCALRCORE-19891
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPolicyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyGroupExists("scalr_policy_group.test", &scalr.PolicyGroup{}),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"name",
						fmt.Sprintf("test-pg-%d", rInt),
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "status"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"error_message",
						"",
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "opa_version"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"account_id",
						defaultAccount,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_provider_id"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.identifier",
						policyGroupVcsRepoID,
					),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.path",
						policyGroupVcsRepoPath,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_repo.0.branch"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "policies.#"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "environments.#"),
				),
			},
		},
	})
}

func TestAccPolicyGroup_update(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// TODO: delete skip after SCALRCORE-19891
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPolicyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyGroupExists("scalr_policy_group.test", &scalr.PolicyGroup{}),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"name",
						fmt.Sprintf("test-pg-%d", rInt),
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "status"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"error_message",
						"",
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "opa_version"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"account_id",
						defaultAccount,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_provider_id"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.identifier",
						policyGroupVcsRepoID,
					),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.path",
						policyGroupVcsRepoPath,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_repo.0.branch"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "policies.#"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "environments.#"),
				),
			},
			{
				Config: testAccPolicyGroupUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyGroupExists("scalr_policy_group.test", &scalr.PolicyGroup{}),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"name",
						"updated_name",
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "status"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"error_message",
						"",
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "opa_version"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"account_id",
						defaultAccount,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_provider_id"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.identifier",
						policyGroupVcsRepoID,
					),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.path",
						policyGroupVcsRepoPath,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_repo.0.branch"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "policies.#"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "environments.#"),
				),
			},
		},
	})
}

func TestAccPolicyGroup_renamed(t *testing.T) {
	rInt := GetRandomInteger()
	policyGroup := &scalr.PolicyGroup{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// TODO: delete skip after SCALRCORE-19891
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPolicyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyGroupExists("scalr_policy_group.test", policyGroup),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"name",
						fmt.Sprintf("test-pg-%d", rInt),
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "status"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"error_message",
						"",
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "opa_version"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"account_id",
						defaultAccount,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_provider_id"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.identifier",
						policyGroupVcsRepoID,
					),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.path",
						policyGroupVcsRepoPath,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_repo.0.branch"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "policies.#"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "environments.#"),
				),
			},
			{
				PreConfig: testAccCheckPolicyGroupRename(policyGroup),
				Config:    testAccPolicyGroupRenamedConfig(rInt),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"name",
						"renamed-outside-of-terraform",
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "status"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"error_message",
						"",
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "opa_version"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"account_id",
						defaultAccount,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_provider_id"),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.identifier",
						policyGroupVcsRepoID,
					),
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"vcs_repo.0.path",
						policyGroupVcsRepoPath,
					),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "vcs_repo.0.branch"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "policies.#"),
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "environments.#"),
				),
			},
		},
	})
}

func TestAccPolicyGroup_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// TODO: delete skip after SCALRCORE-19891
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckPolicyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupBasicConfig(rInt),
			},
			{
				ResourceName:      "scalr_policy_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPolicyGroupExists(resID string, policyGroup *scalr.PolicyGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resID]
		if !ok {
			return fmt.Errorf("not found: %s", resID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		pg, err := scalrClient.PolicyGroups.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*policyGroup = *pg
		return nil
	}
}

func testAccCheckPolicyGroupDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_policy_group" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := scalrClient.PolicyGroups.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("policy group %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckPolicyGroupRename(policyGroup *scalr.PolicyGroup) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		_, err := scalrClient.PolicyGroups.Update(
			context.Background(),
			policyGroup.ID,
			scalr.PolicyGroupUpdateOptions{Name: scalr.String("renamed-outside-of-terraform")},
		)
		if err != nil {
			log.Fatalf("Could not rename policy group outside of terraform: %v", err)
		}
	}
}

func testAccPolicyGroupBasicConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_vcs_provider" "test" {
  name     = "test-github-%d"
  vcs_type = "%s"
  token    = "%s"
}

resource "scalr_policy_group" "test" {
  name            = "test-pg-%[1]d"
  account_id      = "%[4]s"
  vcs_provider_id = scalr_vcs_provider.test.id
  vcs_repo {
	identifier = "%s"
    path       = "%s"
  }
}
`, rInt, string(scalr.Github), githubToken, defaultAccount, policyGroupVcsRepoID, policyGroupVcsRepoPath)
}

func testAccPolicyGroupUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_vcs_provider" "test" {
  name     = "test-github-%d"
  vcs_type = "%s"
  token    = "%s"
}

resource "scalr_policy_group" "test" {
  name            = "updated_name"
  account_id      = "%[4]s"
  vcs_provider_id = scalr_vcs_provider.test.id
  vcs_repo {
	identifier = "%s"
    path       = "%s"
  }
}
`, rInt, string(scalr.Github), githubToken, defaultAccount, policyGroupVcsRepoID, policyGroupVcsRepoPath)
}

func testAccPolicyGroupRenamedConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_vcs_provider" "test" {
  name     = "test-github-%d"
  vcs_type = "%s"
  token    = "%s"
}

resource "scalr_policy_group" "test" {
  name            = "renamed-outside-of-terraform"
  account_id      = "%[4]s"
  vcs_provider_id = scalr_vcs_provider.test.id
  vcs_repo {
	identifier = "%s"
    path       = "%s"
  }
}
`, rInt, string(scalr.Github), githubToken, defaultAccount, policyGroupVcsRepoID, policyGroupVcsRepoPath)
}
