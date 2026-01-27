package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccScalrVcsProviderDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			Steps: []resource.TestStep{
				{
					Config: testAccScalrVcsProviderDataSourceConfigAllFilters(rInt, githubToken),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_vcs_provider.test", "id"),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "name", fmt.Sprintf("vcs-provider-test-%d", rInt),
						),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "vcs_type", "github",
						),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "url", "https://github.com",
						),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "account_id", defaultAccount,
						),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "draft_pr_runs_enabled", "false",
						),
					),
				},
				{
					Config: testAccScalrVcsProviderDataSourceConfigFilterByName(rInt, githubToken),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_vcs_provider.test", "id"),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "name", fmt.Sprintf("vcs-provider-test-%d", rInt),
						),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "vcs_type", "github",
						),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "url", "https://github.com",
						),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "account_id", defaultAccount,
						),
					),
				},
				{
					Config: `
				data scalr_vcs_provider test {
				  vcs_type = "github"
				}`,
					ExpectError: regexp.MustCompile("Your query returned more than one result"),
					PlanOnly:    true,
				},
				{
					Config: `
				data scalr_vcs_provider test {
				  name = "not-existing-vcs"
				}`,
					ExpectError: regexp.MustCompile("Could not find VCS provider matching you query"),
					PlanOnly:    true,
				},
				{
					Config:      `data scalr_vcs_provider test_id {id = ""}`,
					ExpectError: regexp.MustCompile("Attribute id must not be empty"),
					PlanOnly:    true,
				},
				{
					Config:      `data scalr_vcs_provider test_name {name = ""}`,
					ExpectError: regexp.MustCompile("Attribute name must not be empty"),
					PlanOnly:    true,
				},
				// Final step with valid config to allow proper cleanup/destroy
				{
					Config: testAccScalrVcsProviderDataSourceConfigFilterByName(rInt, githubToken),
				},
			},
		},
	)
}

func testAccScalrVcsProviderDataSourceConfigAllFilters(rInt int, token string) string {
	return fmt.Sprintf(
		`
resource scalr_vcs_provider test {
  name       = "vcs-provider-test-%[1]d"
  vcs_type   = "github"
  token      = "%s"
  account_id = "%s"
}

data scalr_vcs_provider test {
  id         = scalr_vcs_provider.test.id
  name       = scalr_vcs_provider.test.name
  vcs_type   = scalr_vcs_provider.test.vcs_type
}`, rInt, token, defaultAccount,
	)
}

func testAccScalrVcsProviderDataSourceConfigFilterByName(rInt int, token string) string {
	return fmt.Sprintf(
		`
resource scalr_vcs_provider test {
  name       = "vcs-provider-test-%[1]d"
  vcs_type   = "github"
  token      = "%s"
  account_id = "%s"
}

data scalr_vcs_provider test {
  name = scalr_vcs_provider.test.name
}`, rInt, token, defaultAccount,
	)
}

func TestAccScalrVcsProviderDataSource_nameMatching(t *testing.T) {
	rInt := GetRandomInteger()
	name1 := fmt.Sprintf("vcs-provider-test-%d", rInt)
	name2 := fmt.Sprintf("vcs-provider-test-%d-suffix", rInt)
	partialName := fmt.Sprintf("test-%d-suffix", rInt)

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			Steps: []resource.TestStep{
				// Exact name match
				{
					Config: testAccScalrVcsProviderDataSourceConfigNameMatching(name1, name2, githubToken, name1),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_vcs_provider.test", "id"),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "name", name1,
						),
					),
				},
				// Partial name match fallback
				{
					Config: testAccScalrVcsProviderDataSourceConfigNameMatching(name1, name2, githubToken, partialName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_vcs_provider.test", "id"),
						resource.TestCheckResourceAttr(
							"data.scalr_vcs_provider.test", "name", name2,
						),
					),
				},
			},
		},
	)
}

func testAccScalrVcsProviderDataSourceConfigNameMatching(name1, name2, token, matchName string) string {
	return fmt.Sprintf(
		`
resource scalr_vcs_provider test1 {
  name       = "%[1]s"
  vcs_type   = "github"
  token      = "%[3]s"
  account_id = "%[4]s"
}

resource scalr_vcs_provider test2 {
  name       = "%[2]s"
  vcs_type   = "github"
  token      = "%[3]s"
  account_id = "%[4]s"
}

data scalr_vcs_provider test {
  name       = "%[5]s"
  depends_on = [scalr_vcs_provider.test1, scalr_vcs_provider.test2]
}`, name1, name2, token, defaultAccount, matchName,
	)
}

func TestAccScalrVcsProviderDataSource_UpgradeFromSDK(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(
		t, resource.TestCase{
			PreCheck: func() { testVcsAccGithubTokenPreCheck(t) },
			Steps: []resource.TestStep{
				{
					ExternalProviders: map[string]resource.ExternalProvider{
						"scalr": {
							Source:            "registry.scalr.io/scalr/scalr",
							VersionConstraint: "<=3.12.0",
						},
					},
					Config: testAccScalrVcsProviderDataSourceConfigUpgradeFromSDK(rInt, githubToken),
				},
				{
					ProtoV5ProviderFactories: protoV5ProviderFactories(t),
					Config:                   testAccScalrVcsProviderDataSourceConfigUpgradeFromSDK(rInt, githubToken),
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

func testAccScalrVcsProviderDataSourceConfigUpgradeFromSDK(rInt int, token string) string {
	return fmt.Sprintf(
		`
resource scalr_vcs_provider test {
  name       = "vcs-provider-test-%[1]d"
  vcs_type   = "github"
  token      = "%s"
  account_id = "%s"
}

data scalr_vcs_provider test {
  name = scalr_vcs_provider.test.name
}

output "test-id" {
  value = data.scalr_vcs_provider.test.id
}

output "test-name" {
  value = data.scalr_vcs_provider.test.name
}`, rInt, token, defaultAccount,
	)
}
