
# Resource `scalr_slack_integration`

Manage the state of slack integrations in Scalr. Create, update and destroy.
Slack workspace should be connected to scalr account before using this resource.

## Example Usage

Basic usage:

```hcl
resource "scalr_slack_integration" "test" {
  name           = "my-channel"
  account_id     = "acc-xxxx"
  events		 = ["run_approval_required", "run_success", "run_errored"]
  channel_id	 = "xxxx" //Can be found in slack UI (channel settings/info popup)
  environments = ["env-xxxxx"]
  workspaces   = ["ws-xxxx", "ws-xxxx"]
}
```

## Argument Reference

* `name` - (Required) Name of the slack integration.
* `channel_id` - (Required) Slack channel event should be sent to.
* `events` - (Required) Terraform run events you would like to receive a Slack notifications for.
Supported values are `run_approval_required`, `run_success`, `run_errored`
* `environments` - (Required) List of environments where events should be triggered on.
* `environments` - (Optional) List of workspaces where events should be triggered on.
Workspaces should be in provided environments.
* `account_id` - (Optional) ID of the account.


## Attribute Reference

All arguments plus:

* `id` - The ID of the Slack integration.
