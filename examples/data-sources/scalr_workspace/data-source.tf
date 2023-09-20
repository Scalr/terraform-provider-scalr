data "scalr_workspace" "example1" {
  id             = "ws-xxxxxxx"
  environment_id = "env-xxxxxxx"
}

data "scalr_workspace" "example2" {
  name           = "my-workspace-name"
  environment_id = "env-xxxxxxx"
}
