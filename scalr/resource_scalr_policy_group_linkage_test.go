package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccPolicyGroupLinkage_basic(t *testing.T) {
	rInt := GetRandomInteger()
	policyGroup := &scalr.PolicyGroup{}
	environment := &scalr.Environment{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testVcsAccGithubTokenPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyGroupLinkageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupLinkageBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyGroupLinkageExists(
						"scalr_policy_group_linkage.test",
						policyGroup,
						environment,
					),
					resource.TestCheckResourceAttrPtr(
						"scalr_policy_group_linkage.test",
						"policy_group_id",
						&policyGroup.ID,
					),
					resource.TestCheckResourceAttrPtr(
						"scalr_policy_group_linkage.test",
						"environment_id",
						&environment.ID,
					),
				),
			},
		},
	})
}

func TestAccPolicyGroupLinkage_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testVcsAccGithubTokenPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPolicyGroupLinkageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupLinkageBasicConfig(rInt),
			},
			{
				ResourceName:      "scalr_policy_group_linkage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPolicyGroupLinkageExists(
	resID string,
	policyGroup *scalr.PolicyGroup,
	environment *scalr.Environment,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resID]
		if !ok {
			return fmt.Errorf("not found: %s", resID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		pg, env, err := getLinkedResources(rs.Primary.ID, scalrClient)
		if err != nil {
			return err
		}

		*policyGroup = *pg
		*environment = *env

		return nil
	}
}

func testAccCheckPolicyGroupLinkageDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_policy_group_linkage" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, _, err := getLinkedResources(rs.Primary.ID, scalrClient)
		if err == nil {
			return fmt.Errorf("policy group linkage %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPolicyGroupLinkageBasicConfig(rInt int) string {
	return fmt.Sprintf(`
locals {
  account_id = "%s"
}

resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = local.account_id
}

resource "scalr_vcs_provider" "test" {
  name     = "test-github-%[2]d"
  vcs_type = "%s"
  token    = "%s"
}

resource "scalr_policy_group" "test" {
  name            = "test-pg-%[2]d"
  account_id      = local.account_id
  vcs_provider_id = scalr_vcs_provider.test.id
  vcs_repo {
	identifier = "%[5]s"
    path       = "%s"
  }
}

resource "scalr_policy_group_linkage" "test" {
  policy_group_id = scalr_policy_group.test.id
  environment_id  = scalr_environment.test.id
}
`, defaultAccount, rInt, string(scalr.Github), GITHUB_TOKEN, policyGroupVcsRepoID, policyGroupVcsRepoPath)
}
