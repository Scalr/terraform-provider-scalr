
# Resource `scalr_workspace_run_schedule` 

Allows workspace admins to automate the configuration of recurring runs for a workspace.

## Example Usage

Basic usage:

```hcl
data scalr_environment "current" {
  account_id = "acc-12345"
  name = "dev"
}

data "scalr_workspace" "cert" {
  environment_id = data.scalr_environment.current.id
  name = "ssl-certificates"
}

resource "scalr_workspace_run_schedule" "example" {
  workspace_id = data.scalr_workspace.cert.id
  apply_schedule = "30 3 5 3-5 2"
  destroy_schedule = "30 4 5 3-5 2"
}
```

## Argument Reference

* `workspace_id` - (Required) ID of the workspace, in the format `ws-<RANDOM STRING>`.
* `apply_schedule` - (Optional) Cron expression for when apply run should be created.
* `destroy_schedule` - (Optional) Cron expression for when destroy run should be created.


## Attribute Reference

All arguments plus:

* `id` - The identifier of a workspace in the format `ws-<RANDOM STRING>`.

