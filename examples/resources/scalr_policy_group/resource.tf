resource "scalr_policy_group" "example" {
  name            = "instance_types"
  opa_version     = "0.29.4"
  account_id      = "acc-xxxxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
    path       = "policies/instance"
    branch     = "dev"
  }
}

# Pinned to a Git tag: pushes to the repository do not change the policies
# until the tag is replaced with another one.
resource "scalr_policy_group" "pinned" {
  name            = "instance_types_pinned"
  account_id      = "acc-xxxxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
    path       = "policies/instance"
    tag        = "v1.4.0"
  }
}
