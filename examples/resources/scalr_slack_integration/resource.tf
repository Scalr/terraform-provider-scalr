resource "scalr_slack_integration" "test" {
  name         = "my-channel"
  account_id   = "acc-xxxx"
  events       = ["run_approval_required", "run_success", "run_errored"]
  channel_id   = "xxxx" # Can be found in slack UI (channel settings/info popup)
  environments = ["env-xxxxx"]
  workspaces   = ["ws-xxxx", "ws-xxxx"]
}
