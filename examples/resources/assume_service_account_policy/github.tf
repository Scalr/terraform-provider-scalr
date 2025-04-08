resource "scalr_workload_identity_provider" "github" {
  name              = "github-actions"
  url               = "https://token.actions.githubusercontent.com"
  allowed_audiences = ["scalr-github-actions"]
}

resource "scalr_assume_service_account_policy" "ga-scalr-staging" {
  name                     = "ga-scalr-staging"
  service_account_id       = scalr_service_account.staging.id
  provider_id              = scalr_workload_identity_provider.github.id
  maximum_session_duration = 7200
  claim_condition {
    claim    = "sub"
    value    = "repo:GithubOrganization/repository:environment:staging"
    operator = "startswith"
  }
}
