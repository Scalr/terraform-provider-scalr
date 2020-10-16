---
layout: "scalr"
page_title: "Scalr: scalr_variable"
sidebar_current: "docs-resource-scalr-variable"
description: |-
  Manages variables.
---

# scalr_variable

Manage the state of variables in Scalr. Creates, updates and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_workspace" "test" {
  name           = "my-workspace-name"
  environment_id = "env-xxxxxxxxxx"
}

resource "scalr_variable" "test" {
  key          = "my_key_name"
  value        = "my_value_name"
  category     = "terraform"
  workspace_id = scalr_workspace.test.id
}
```

## Arguments

* `key` - (Required) Name of the variable.
* `value` - (Required) Value of the variable.
* `category` - (Required) Indicates if this is a Terraform or environment variable. Allowed values are `terraform` or `env`.
* `hcl` - (Optional) Set (true/false) to configure if the value of the variable as a string of HCL code. Has no effect for `category = "env"` variables. Defaults to `false`.
* `sensitive` - (Optional) Set (true/false) to configure if the value is sensitive. Sensitive variable values are not visible after being set. Defaults to `false`.
* `workspace_id` - (Required) The workspace that owns the variable, specified as
  an ID, in the format `ws-<RANDOM STRING>`.

## Attributes

All arguments plus:

* `id` - The ID of the variable, in the format `var-<RANDOM STRING>`.

## Import

Variables can be imported; use
`<ORGANIZATION NAME>/<WORKSPACE NAME>/<VARIABLE ID>` as the import ID. For
example:

```shell
terraform import scalr_variable.test my-org-name/my-workspace-name/var-5rTwnSaRPogw6apb
```
