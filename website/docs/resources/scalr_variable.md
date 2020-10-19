---
layout: "scalr"
page_title: "Scalr: scalr_variable"
sidebar_current: "docs-resource-scalr-variable"
description: |-
  Manages variables.
---

# scalr_variable Resource

Manage the state of variables in Scalr. Creates, updates and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_workspace" "example" {
  name           = "my-workspace-name"
  environment_id = "env-xxxxxxxxxx"
}

resource "scalr_variable" "example" {
  key          = "my_key_name"
  value        = "my_value_name"
  category     = "terraform"
  workspace_id = scalr_workspace.example.id
}
```

## Argument Reference

* `key` - (Required) Key of the variable.
* `value` - (Required) Variable value.
* `category` - (Required) Indicates if this is a Terraform or environment variable. Allowed values are `terraform` or `env`.
* `hcl` - (Optional) Set (true/false) to configure the variable as a string of HCL code. Has no effect for `category = "env"` variables. Default `false`.
* `sensitive` - (Optional) Set (true/false) to configure as sensitive. Sensitive variable values are not visible after being set. Default `false`.
* `workspace_id` - (Required) The workspace that owns the variable, specified as an ID, in the format `ws-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - The ID of the variable, in the format `var-<RANDOM STRING>`.

## Import

Variables can be imported; use
`<Environment NAME>/<WORKSPACE NAME>/<VARIABLE ID>` as the import ID. For
example:

```shell
terraform import scalr_variable.example environment-name/workspace-name/var-xxxxxxxxxxxx
```
