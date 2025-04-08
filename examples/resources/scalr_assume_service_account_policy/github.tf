data "scalr_workload_identity_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
}

resource "scalr_assume_service_account_policy" "ga-scalr-staging" {
  name                     = "ga-scalr-staging"
  service_account_id       = scalr_service_account.staging.id
  provider_id              = data.scalr_workload_identity_provider.github.id
  maximum_session_duration = 7200
  claim_condition {
    claim    = "sub"
    value    = "repo:GithubOrganization/repository:environment:staging"
    operator = "startswith"
  }
  claim_condition {
    claim    = "repository"
    value    = "GithubOrganization/repository"
    operator = "eq"
  }
}
