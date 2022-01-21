---
layout: "scalr"
page_title: "Scalr: scalr_run_trigger"
sidebar_current: "docs-resource-scalr-run-trigger"
description: |-
  Manages vcs provider.
---

# scalr_vcs_provider Resource

Run triggers are a way to chain workspaces together. 
The use case for this is that you might have one or more upstream workspaces that need to automatically kick off a downstream workspace based on a successful run in the upstream workspace. 
To set a trigger, go to the downstream workspace and set the upstream workspace(s). 
Now, whenever the upstream workspace has a successful run, the downstream workspace will automatically start a run. 

## Example Usage

Basic usage:

```hcl

data "scalr_workspace" "downstream" {
  name           = "downstream"
  environment_id = "env-xxxxxxxxx"
}

data "scalr_workspace" "upstream" {
  name           = "upstream"
  environment_id = "env-xxxxxxxxx"
}

resource "scalr_run_trigger" "set_downstream" {
   downstream_id  = data.scalr_workspace.downstream.id # run automatically queued in this workspace once the run in upstream workspace is applied
   upstream_id = data.scalr_workspace.upstream.id
}
```

## Argument Reference

* `downstream_id` - (Required) The identifier of the workspace in which new runs will be created.
* `upstream_id` (Required) The identifier of the upstream workspace to track new runs.


## Attribute Reference

All arguments plus:

* `id` - The identifier of created trigger

## Import

To import existing run trigger use its identifier. For example:
```shell
terraform import scalr_run_trigger.set_downstream rt-xxxxxxxxxx
```
