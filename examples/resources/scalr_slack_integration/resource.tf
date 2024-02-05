resource "scalr_slack_integration" "test" {
  name         = "my-channel"
  account_id   = "acc-xxxxxxxxxx"
  events       = ["run_approval_required", "run_success", "run_errored"]
  run_mode     = "apply"
  channel_id   = "xxxxxxxxxx" # Can be found in slack UI (channel settings/info popup)
  environments = ["env-xxxxxxxxxx"]
  workspaces   = ["ws-xxxxxxxxxx", "ws-yyyyyyyyyy"]
}
