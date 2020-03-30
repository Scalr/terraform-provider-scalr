---
layout: "scalr"
page_title: "Scalr: scalr_variable"
sidebar_current: "docs-resource-scalr-variable"
description: |-
  Manages variables.
---

# scalr_variable

Creates, updates and destroys variables.

## Example Usage

Basic usage:

```hcl
resource "scalr_workspace" "test" {
  name         = "my-workspace-name"
  organization = "my-org"
}

resource "scalr_variable" "test" {
  key          = "my_key_name"
  value        = "my_value_name"
  category     = "terraform"
  workspace_id = "${scalr_workspace.test.id}"
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required) Name of the variable.
* `value` - (Required) Value of the variable.
* `category` - (Required) Whether this is a Terraform or environment variable.
  Valid values are `terraform` or `env`.
* `hcl` - (Optional) Whether to evaluate the value of the variable as a string
  of HCL code. Has no effect for environment variables. Defaults to `false`.
* `sensitive` - (Optional) Whether the value is sensitive. If true then the
  variable is written once and not visible thereafter. Defaults to `false`.
* `workspace_id` - (Required) The workspace that owns the variable, specified as
  a human-readable ID (`<ORGANIZATION>/<WORKSPACE>`).

## Attributes Reference

* `id` - The ID of the variable.

## Import

Variables can be imported; use
`<ORGANIZATION NAME>/<WORKSPACE NAME>/<VARIABLE ID>` as the import ID. For
example:

```shell
terraform import scalr_variable.test my-org-name/my-workspace-name/var-5rTwnSaRPogw6apb
```
