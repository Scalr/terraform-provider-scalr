resource "scalr_workspace_run_schedule" "example" {
  workspace_id     = "ws-xxxxxxxxxx"
  apply_schedule   = "30 3 5 3-5 2"
  destroy_schedule = "30 4 5 3-5 2"
}
