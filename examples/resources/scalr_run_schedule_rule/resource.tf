resource "scalr_run_schedule_rule" "example" {
  schedule      = "0 4 * * *"
  schedule_mode = "apply"
  workspace_id  = "ws-xxxxxxxxxx"
}
