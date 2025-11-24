resource "scalr_drift_detection" "example" {
  environment_id = "env-xxxxx"
  check_period   = "weekly"
}

resource "scalr_drift_detection" "example" {
  environment_id = "env-xxxxx"
  check_period   = "weekly"
  run_mode       = "plan"
  workspace_filters {
    name_patterns = ["prod", "stage-*"]
  }
}
