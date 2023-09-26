data "scalr_workspace" "example1" {
  id             = "ws-xxxxxxxxxx"
  environment_id = "env-xxxxxxxxxx"
}

data "scalr_workspace" "example2" {
  name           = "my-workspace-name"
  environment_id = "env-xxxxxxxxxx"
}
