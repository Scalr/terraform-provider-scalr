---
layout: "scalr"
page_title: "Scalr: scalr_role"
sidebar_current: "docs-datasource-scalr-role-x"
description: |-
  Get information on a IAM role.
---

# scalr_role Data Source

This data source is used to retrieve details of a single role by name and account_id.

## Example Usage

```hcl
data "scalr_role" "example" {
  name           = "WorkspaceAdmin"
  account_id     = "acc-xxxxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the role.
* `account_id` - (Required) ID of the account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - The ID of the role.
* `permissions` - Array of permission names.
* `is_system` - Boolean indicates if the role can be edited.
* `description` - Verbose description of the role.
