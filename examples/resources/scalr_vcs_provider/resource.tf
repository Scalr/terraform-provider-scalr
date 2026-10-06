resource "scalr_vcs_provider" "example" {
  name       = "example-github"
  account_id = "acc-xxxxxxxxxx"
  vcs_type   = "github"
  token      = "token"
}

resource "scalr_vcs_provider" "azure_dev_ops" {
  name       = "example-azure-devops"
  account_id = "acc-xxxxxxxxxx"
  vcs_type   = "azure_dev_ops_services"
  token      = "token"
  username   = "my-organization" # required only for organization-scoped tokens
}
