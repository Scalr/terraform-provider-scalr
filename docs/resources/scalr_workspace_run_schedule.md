---
layout: "scalr"
page_title: "Scalr: scalr_workspace_run_schedule"
sidebar_current: "docs-resource-scalr-workspace-run-schedule"
description: |-
Manages run schedules.
---

# scalr_workspace_run_schedule Resource

Manage the state of workspace run schedules in Scalr. Create, update and destroy

## Example Usage

Basic usage:

```hcl
resource "scalr_workspace_run_schedule" "example" {
  workspace_id = "ws-xxxxxx"
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

* `id` - The workspaces's ID, in the format `ws-<RANDOM STRING>`.

