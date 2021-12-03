---
layout: "scalr"
page_title: "Scalr: scalr_agent_pool"
sidebar_current: "docs-datasource-scalr-agent-pool"
description: |-
  Get information on the agent pool.
---

# scalr_agent_pool Data Source

This data source is used to retrieve details of an agent pool.

## Example Usage

Basic usage:

```hcl
data "scalr_agent_pool" "default" {
  name       = "default-pool"
  account_id = "acc-xxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) Name of the agent pool.
* `account_id` - (Required) ID of the account.
* `environment_id` - (Optional) ID of the environment.

## Attribute Reference

All arguments plus:

* `id` - The ID of the agent pool.
* `workspace_ids` - The list of IDs of linked workspaces.