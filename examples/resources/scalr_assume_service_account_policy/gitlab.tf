data "scalr_workload_identity_provider" "gitlab" {
  url = "https://gitlab.com"
}

resource "scalr_assume_service_account_policy" "gitlab-ci-scalr-staging" {
  name                     = "gitlab-ci-scalr-staging"
  service_account_id       = scalr_service_account.staging.id
  provider_id              = data.scalr_workload_identity_provider.gitlab.id
  maximum_session_duration = 3600
  claim_condition {
    claim    = "sub"
    value    = "group/project:ref_type:type:ref:branch_name"
    operator = "eq"
  }
}
