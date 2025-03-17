resource "scalr_environment_hook" "test_link" {
  hook_id        = "hook-xxxxx"
  environment_id = "env-xxxxx"
  events         = ["pre-init", "post-appy"]
}

resource "scalr_environment_hook" "test_link_all" {
  hook_id        = "hook-xxxxx"
  environment_id = "env-xxxxx"
  events         = ["*"]
}