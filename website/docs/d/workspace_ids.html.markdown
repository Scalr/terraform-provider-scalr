---
layout: "scalr"
page_title: "Scalr: scalr_workspace_ids"
sidebar_current: "docs-datasource-scalr-workspace-ids"
description: |-
  Get information on workspace IDs.
---

# scalr_workspace_ids

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

## Arguments

* `names` - (Required) A list of workspace names to search for. Names that don't
  match a real workspace will be omitted from the results, but are not an error.

    To select _all_ workspaces for an environment, provide a list with a single
    asterisk, like `["*"]`. No other use of wildcards is supported.
* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.

## Attributes

All arguments plus:

* `ids` - A map of workspace names and their opaque IDs, in the format `ws-<RANDOM STRING>`.
