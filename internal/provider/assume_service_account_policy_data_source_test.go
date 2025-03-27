package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var aspName = acctest.RandomWithPrefix("test-asp")

func TestAccScalrAssumeServiceAccountPolicyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAssumeServiceAccountPolicyDataSourceInitConfig,
			},
			{
				Config: testAccScalrAssumeServiceAccountPolicyDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_assume_service_account_policy.github", "scalr_assume_service_account_policy.github"),
					testAccCheckEqualID("data.scalr_assume_service_account_policy.gitlab", "scalr_assume_service_account_policy.gitlab"),
					testAccCheckEqualID("data.scalr_assume_service_account_policy.bitbucket", "scalr_assume_service_account_policy.bitbucket"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.github", "name",
						"scalr_assume_service_account_policy.github", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.github", "service_account_id",
						"scalr_assume_service_account_policy.github", "service_account_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.github", "provider_id",
						"scalr_assume_service_account_policy.github", "provider_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.github", "maximum_session_duration",
						"scalr_assume_service_account_policy.github", "maximum_session_duration",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.gitlab", "name",
						"scalr_assume_service_account_policy.gitlab", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.bitbucket", "name",
						"scalr_assume_service_account_policy.bitbucket", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.github", "claim_conditions.0.claim",
						"scalr_assume_service_account_policy.github", "claim_condition.0.claim",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.github", "claim_conditions.0.value",
						"scalr_assume_service_account_policy.github", "claim_condition.0.value",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_assume_service_account_policy.github", "claim_conditions.0.operator",
						"scalr_assume_service_account_policy.github", "claim_condition.0.operator",
					),
				),
			},
			{
				Config: testAccScalrAssumeServiceAccountPolicyDataSourceInitConfig,
			},
		},
	})
}

var testAccScalrAssumeServiceAccountPolicyDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_service_account" "github" {
  name = "%[1]s-github"
}

resource "scalr_service_account" "gitlab" {
  name = "%[1]s-gitlab"
}

resource "scalr_service_account" "bitbucket" {
  name = "%[1]s-bitbucket"
}

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

resource "scalr_workload_identity_provider" "bitbucket" {
  name               = "%[1]s-bitbucket"
  url                = "https://bitbucket.com"
  allowed_audiences  = ["bitbucketscalr"]
}

resource "scalr_assume_service_account_policy" "github" {
  name                     = "%[1]s-github"
  service_account_id       = scalr_service_account.github.id
  provider_id              = scalr_workload_identity_provider.github.id
  maximum_session_duration = 4000
  claim_condition {
    claim    = "sub"
    value    = "12345"
    operator = "eq"
  }
}

resource "scalr_assume_service_account_policy" "gitlab" {
  name                     = "%[1]s-gitlab"
  service_account_id       = scalr_service_account.gitlab.id
  provider_id              = scalr_workload_identity_provider.gitlab.id
  claim_condition {
    claim    = "sub"
    value    = "67890"
  }
}

resource "scalr_assume_service_account_policy" "bitbucket" {
  name                     = "%[1]s-bitbucket"
  service_account_id       = scalr_service_account.bitbucket.id
  provider_id              = scalr_workload_identity_provider.bitbucket.id
  claim_condition {
    claim    = "sub"
    value    = "67890"
  }
}
`, aspName)

var testAccScalrAssumeServiceAccountPolicyDataSourceConfig = testAccScalrAssumeServiceAccountPolicyDataSourceInitConfig + `
data "scalr_assume_service_account_policy" "github" {
  name               = scalr_assume_service_account_policy.github.name
  service_account_id = scalr_service_account.github.id
}

data "scalr_assume_service_account_policy" "gitlab" {
  name               = scalr_assume_service_account_policy.gitlab.name
  service_account_id = scalr_service_account.gitlab.id
}

data "scalr_assume_service_account_policy" "bitbucket" {
  id                 = scalr_assume_service_account_policy.bitbucket.id
  service_account_id = scalr_service_account.bitbucket.id
}
`
