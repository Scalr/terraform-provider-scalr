package provider

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"

	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"
)

const (
	policyGroupVcsRepoID   = "Scalr/tf-revizor-fixtures"
	policyGroupVcsRepoPath = "policies/clouds"
	commonFunctionsFolder  = "policies/instances"
)

func TestAccPolicyGroup_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// TODO: delete skip after SCALRCORE-19891
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckPolicyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyGroupExists("scalr_policy_group.test", &schemas.PolicyGroup{}),
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
					resource.TestCheckResourceAttr("scalr_policy_group.test", "execution_mode", "post-plan"),
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
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"common_functions_folder",
						commonFunctionsFolder,
					),
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
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckPolicyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyGroupExists("scalr_policy_group.test", &schemas.PolicyGroup{}),
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
					resource.TestCheckResourceAttr("scalr_policy_group.test", "execution_mode", "post-plan"),
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
					resource.TestCheckResourceAttr(
						"scalr_policy_group.test",
						"common_functions_folder",
						commonFunctionsFolder,
					),
				),
			},
			{
				Config: testAccPolicyGroupUpdateConfig(rInt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("scalr_policy_group.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyGroupExists("scalr_policy_group.test", &schemas.PolicyGroup{}),
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
					resource.TestCheckResourceAttr("scalr_policy_group.test", "execution_mode", "pre-plan"),
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
					resource.TestCheckResourceAttr("scalr_policy_group.test", "common_functions_folder", ""),
				),
			},
		},
	})
}

func TestAccPolicyGroup_renamed(t *testing.T) {
	rInt := GetRandomInteger()
	policyGroup := &schemas.PolicyGroup{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// TODO: delete skip after SCALRCORE-19891
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckPolicyGroupDestroy,
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
					resource.TestCheckResourceAttr("scalr_policy_group.test", "execution_mode", "post-plan"),
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
				// Rename the policy group outside of Terraform and refresh the state.
				// The plan still has a diff, because the config keeps the old name.
				PreConfig:          testAccCheckPolicyGroupRename(policyGroup),
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
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
					resource.TestCheckResourceAttr("scalr_policy_group.test", "execution_mode", "post-plan"),
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
				// The config matching the new name produces no changes.
				Config:   testAccPolicyGroupRenamedConfig(rInt),
				PlanOnly: true,
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
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckPolicyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupBasicConfig(rInt),
			},
			{
				ResourceName:      "scalr_policy_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The policy group is still being fetched from VCS right after creation,
				// so its status and policies differ by the time it is imported.
				ImportStateVerifyIgnore: []string{"status", "policies"},
			},
		},
	})
}

func testAccCheckPolicyGroupExists(resID string, policyGroup *schemas.PolicyGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resID]
		if !ok {
			return fmt.Errorf("not found: %s", resID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		pg, err := scalrClient.PolicyGroup.GetPolicyGroup(ctx, rs.Primary.ID, nil)
		if err != nil {
			return err
		}

		*policyGroup = *pg
		return nil
	}
}

func testAccCheckPolicyGroupDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_policy_group" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := scalrClient.PolicyGroup.GetPolicyGroup(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("policy group %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckPolicyGroupRename(policyGroup *schemas.PolicyGroup) func() {
	return func() {
		scalrClient := createScalrClientV2()

		_, err := scalrClient.PolicyGroup.UpdatePolicyGroup(
			context.Background(),
			policyGroup.ID,
			&schemas.PolicyGroupRequest{
				Attributes: schemas.PolicyGroupAttributesRequest{
					Name: value.Set("renamed-outside-of-terraform"),
				},
			},
			nil,
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
  common_functions_folder = "%s"
}
`,
		rInt,
		string(scalr.Github),
		githubToken,
		defaultAccount,
		policyGroupVcsRepoID,
		policyGroupVcsRepoPath,
		commonFunctionsFolder,
	)
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
  execution_mode  = "pre-plan"
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
  common_functions_folder = "%s"
}
`,
		rInt,
		string(scalr.Github),
		githubToken,
		defaultAccount,
		policyGroupVcsRepoID,
		policyGroupVcsRepoPath,
		commonFunctionsFolder,
	)
}
