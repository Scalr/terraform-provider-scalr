resource "scalr_workspace" "parent" {
  name                 = "parent"
  environment_id       = "env-xxxxxxxxxx"
  remote_state_sharing = false
}

resource "scalr_workspace" "child" {
  name           = "child"
  environment_id = "env-xxxxxxxxxx"
}

resource "scalr_workspace_remote_state_consumer" "child" {
  workspace_id = scalr_workspace.parent.id
  consumer_id  = scalr_workspace.child.id
}
