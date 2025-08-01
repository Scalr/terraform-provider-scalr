---
title: "scalr_slack_integration"
categorySlug: "scalr-terraform-provider"
slug: "provider_resource_scalr_slack_integration"
parentDocSlug: "provider_resources"
hidden: false
order: 23
---
## Resource: scalr_slack_integration

Manage the state of Slack integrations in Scalr. Create, update and destroy.

-> **Note** Slack workspace should be connected to Scalr account before using this resource.

## Example Usage

```terraform
resource "scalr_slack_integration" "test" {
  name         = "my-channel"
  account_id   = "acc-xxxxxxxxxx"
  events       = ["run_approval_required", "run_success", "run_errored"]
  run_mode     = "apply"
  channel_id   = "xxxxxxxxxx" # Can be found in slack UI (channel settings/info popup)
  environments = ["env-xxxxxxxxxx"]
  workspaces   = ["ws-xxxxxxxxxx", "ws-yyyyyyyyyy"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `channel_id` (String) Slack channel ID the event will be sent to.
- `environments` (Set of String) List of environments where events should be triggered.
- `events` (Set of String) Terraform run events you would like to receive a Slack notifications for. Supported values are `run_approval_required`, `run_success`, `run_errored`.
- `name` (String) Name of the Slack integration.

### Optional

- `account_id` (String) ID of the account.
- `run_mode` (String) What type of runs should be reported, available options: `all`, `apply`, `dry`.
- `workspaces` (Set of String) List of workspaces where events should be triggered. Workspaces should be in provided environments. If no workspace is given for a specified environment, events will trigger in all of its workspaces.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
terraform import scalr_slack_integration.example in-xxxxxxxxxx
```
