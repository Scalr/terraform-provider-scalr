data "scalr_environment" "example" {
  name       = "env-name"
  account_id = "acc-xxxxxxxxxx"
}

resource "scalr_workspace" "example" {
  name              = "my-workspace-name"
  environment_id    = data.scalr_environment.example.id
  working_directory = "example/path"
}
