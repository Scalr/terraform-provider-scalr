resource "scalr_vcs_provider" "example" {
  name       = "example-github"
  account_id = "acc-xxxxxxxxxx"
  vcs_type   = "github"
  token      = "token"
}
