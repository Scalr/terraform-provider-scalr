---
layout: "scalr"
page_title: "Scalr: scalr_workspace_ids"
sidebar_current: "docs-datasource-scalr-workspace-ids"
description: |-
  Get information on workspace IDs.
---

# scalr_workspace_ids Data Source

Obtain a map of workspace IDs based on the names provided. Wildcards are accepted.

## Example Usage

```hcl
data "scalr_workspace_ids" "app-frontend" {
  names          = ["app-frontend-prod", "app-frontend-dev1", "app-frontend-staging"]
  environment_id = "env-xxxxxxxxxxx"
}

data "scalr_workspace_ids" "all" {
  names          = ["*"]
  environment_id = "env-xxxxxxxxxxx"
}
```

## Argument Reference

* `names` - (Required)   * A list of names to search for. If a name does not exist, it will not throw an error, it will just not exist in the returned output. Use `["*"]` to select all workspaces.
* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `ids` - A map of workspace names and their opaque IDs, in the format `env_id/name`.
