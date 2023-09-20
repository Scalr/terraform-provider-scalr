resource "scalr_policy_group" "example" {
  name            = "instance_types"
  opa_version     = "0.29.4"
  account_id      = "acc-xxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
    path       = "policies/instance"
    branch     = "dev"
  }
}
