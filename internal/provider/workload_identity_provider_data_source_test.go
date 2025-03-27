package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var wipName = acctest.RandomWithPrefix("test-wip")

func TestAccScalrWorkloadIdentityProviderDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_workload_identity_provider test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_workload_identity_provider test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_workload_identity_provider test {url = ""}`,
				ExpectError: regexp.MustCompile("Attribute url must not be empty"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrWorkloadIdentityProviderDataSourceInitConfig,
			},
			{
				Config: testAccScalrWorkloadIdentityProviderDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_workload_identity_provider.github", "scalr_workload_identity_provider.github"),
					testAccCheckEqualID("data.scalr_workload_identity_provider.gitlab", "scalr_workload_identity_provider.gitlab"),
					testAccCheckEqualID("data.scalr_workload_identity_provider.gitlab_id", "scalr_workload_identity_provider.gitlab"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_workload_identity_provider.github", "name",
						"scalr_workload_identity_provider.github", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_workload_identity_provider.github", "url",
						"scalr_workload_identity_provider.github", "url",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_workload_identity_provider.github", "allowed_audiences",
						"scalr_workload_identity_provider.github", "allowed_audiences",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_workload_identity_provider.gitlab", "name",
						"scalr_workload_identity_provider.gitlab", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_workload_identity_provider.gitlab_id", "name",
						"scalr_workload_identity_provider.gitlab", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_workload_identity_provider.github", "assume_service_account_policies.0",
						"scalr_assume_service_account_policy.github", "id",
					),
					resource.TestCheckResourceAttr("data.scalr_workload_identity_provider.github", "assume_service_account_policies.#", "1"),
				),
			},
			{
				Config: testAccScalrWorkloadIdentityProviderDataSourceInitConfig,
			},
		},
	})
}

var testAccScalrWorkloadIdentityProviderDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_workload_identity_provider" "github" {
  name               = "%[1]s-github"
  url                = "https://token.actions.githubusercontent.com"
  allowed_audiences  = ["githubscalr"]
}

resource "scalr_workload_identity_provider" "gitlab" {
  name               = "%[1]s-gitlab"
  url                = "https://gitlab.com"
  allowed_audiences  = ["gitlabscalr"]
}

resource "scalr_service_account" "github" {
  name = "%[1]s"
}

resource "scalr_assume_service_account_policy" "github" {
  name                     = "%[1]s"
  service_account_id       = scalr_service_account.github.id
  provider_id              = scalr_workload_identity_provider.github.id
  claim_condition {
    claim    = "sub"
    value    = "12345"
    operator = "eq"
  }
}
`, wipName)

var testAccScalrWorkloadIdentityProviderDataSourceConfig = testAccScalrWorkloadIdentityProviderDataSourceInitConfig + `
data "scalr_workload_identity_provider" "github" {
  name = scalr_workload_identity_provider.github.name
}

data "scalr_workload_identity_provider" "gitlab" {
  url = scalr_workload_identity_provider.gitlab.url
}

data "scalr_workload_identity_provider" "gitlab_id" {
  id = scalr_workload_identity_provider.gitlab.id
}
`
