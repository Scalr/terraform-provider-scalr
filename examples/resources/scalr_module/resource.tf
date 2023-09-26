resource "scalr_module" "example" {
  account_id      = "acc-xxxxxxxxxx"
  environment_id  = "env-xxxxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
    path       = "example/terraform-<provider>-<name>"
    tag_prefix = "aws/"
  }
}
