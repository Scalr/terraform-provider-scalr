package scalr

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scalr/go-scalr"
)

func TestAccPolicyGroupDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// TODO: delete skip after SCALRCORE-19891
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyGroupConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_policy_group.test", "id"),
				),
			},
			{
				Config:      `data "scalr_policy_group" "test" {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				PreConfig: waitForPolicyGroupFetch(fmt.Sprintf("test-pg-%d", rInt)),
				Config:    testAccPolicyGroupDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_policy_group.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_policy_group.test",
						"name",
						fmt.Sprintf("test-pg-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"data.scalr_policy_group.test",
						"status",
						string(scalr.PolicyGroupStatusActive),
					),
					resource.TestCheckResourceAttr(
						"data.scalr_policy_group.test",
						"error_message",
						"",
					),
					resource.TestCheckResourceAttrSet("data.scalr_policy_group.test", "opa_version"),
					resource.TestCheckResourceAttr(
						"data.scalr_policy_group.test",
						"account_id",
						defaultAccount,
					),
					resource.TestCheckResourceAttrSet("data.scalr_policy_group.test", "vcs_provider_id"),
					resource.TestCheckResourceAttr(
						"data.scalr_policy_group.test",
						"vcs_repo.0.identifier",
						policyGroupVcsRepoID,
					),
					resource.TestCheckResourceAttr(
						"data.scalr_policy_group.test",
						"vcs_repo.0.path",
						policyGroupVcsRepoPath,
					),
					resource.TestCheckResourceAttrSet("data.scalr_policy_group.test", "vcs_repo.0.branch"),
					resource.TestCheckResourceAttrSet("data.scalr_policy_group.test", "policies.#"),
					resource.TestCheckResourceAttrSet("data.scalr_policy_group.test", "environments.#"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scalr_policy_group" "test" {
					  name       = "not-exists"
					  account_id = "%s"
					}
				`, defaultAccount),
				ExpectError: regexp.MustCompile(fmt.Sprintf(
					"policy group %s/%s not found", defaultAccount, "not-exists",
				)),
				PlanOnly: true,
			},
		},
	})
}

func waitForPolicyGroupFetch(name string) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		pgl, err := scalrClient.PolicyGroups.List(ctx, scalr.PolicyGroupListOptions{
			Account: defaultAccount,
			Name:    name,
		})
		if err != nil || len(pgl.Items) == 0 {
			log.Fatalf("The test policy group on account %s was not created: %v", defaultAccount, err)
		}

		var pgID = pgl.Items[0].ID

		for i := 0; i < 60; i++ {
			pg, err := scalrClient.PolicyGroups.Read(ctx, pgID)
			if err != nil {
				log.Fatalf("Error polling policy group %s: %v", pgID, err)
			}
			if pg.Status != scalr.PolicyGroupStatusFetching {
				if pg.Status != scalr.PolicyGroupStatusActive {
					log.Fatalf("Invalid policy group status: '%s'", pg.Status)
				}
				return
			}
			time.Sleep(time.Second)
		}
		log.Fatal("Policy group has not become active after 60 seconds")
	}
}

func testAccPolicyGroupConfig(rInt int) string {
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

func testAccPolicyGroupDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
%s

data "scalr_policy_group" "test" {
  id         = scalr_policy_group.test.id
  name       = scalr_policy_group.test.name
  account_id = "%s"
}
`, testAccPolicyGroupConfig(rInt), defaultAccount)
}
