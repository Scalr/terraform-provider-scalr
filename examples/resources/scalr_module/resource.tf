resource "scalr_module" "example" {
  account_id      = "acc-xxxxxxxxx"
  environment_id  = "env-xxxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
    path       = "example/terraform-<provider>-<name>"
    tag_prefix = "aws/"
  }
}
