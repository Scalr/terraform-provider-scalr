resource "scalr_module_test_configuration" "example" {
  module_id                      = scalr_module.example.id
  enabled                        = true
  failure_behavior               = "notify"
  trigger_on_pr_activity_enabled = true
  trigger_on_new_version_enabled = true
}
