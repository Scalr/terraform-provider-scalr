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
  description  = "variable description"
  workspace_id = scalr_workspace.example.id
}
```

Adding Complex Type Variable:

 ```javascript
 resource "scalr_variable" "example" {
   key          = "xyz"
   value        = jsonencode(["foo", "bar"])
   hcl          = true
   category     = "terraform"
   workspace_id = scalr_workspace.example.id
 }
 ```

## Argument Reference

* `key` - (Required) Key of the variable.
* `value` - (Required) Variable value.
* `category` - (Required) Indicates if this is a Terraform or shell variable. Allowed values are `terraform` or `shell`.
* `description` - (Optional) Variable verbose description, defaults to empty string.
* `hcl` - (Optional) Set (true/false) to configure the variable as a string of HCL code. Has no effect for `category = "shell"` variables. Default `false`.
* `sensitive` - (Optional) Set (true/false) to configure as sensitive. Sensitive variable values are not visible after being set. Default `false`.
* `final` - (Optional) Set (true/false) to configure as final. Indicates whether the variable can be overridden on a lower scope down the Scalr organizational model. Default `false`.
* `force` - (Optional) Set (true/false) to configure as force. Allows creating final variables on higher scope, even if the same variable exists on lower scope (lower is to be deleted). Default `false`.
* `workspace_id` - (Optional) The workspace that owns the variable, specified as an ID, in the format `ws-<RANDOM STRING>`.
* `environment_id` - (Optional) The environment that owns the variable, specified as an ID, in the format `env-<RANDOM STRING>`.
* `account_id` - (Optional) The account that owns the variable, specified as an ID, in the format `acc-<RANDOM STRING>`.


## Attribute Reference

All arguments plus:

* `id` - The ID of the variable, in the format `var-<RANDOM STRING>`.

## Import

To import variables use `<Environment NAME>/<WORKSPACE NAME>/<VARIABLE ID>` as the import ID. For example:

```shell
terraform import scalr_variable.example environment-name/workspace-name/var-xxxxxxxxxxxx
```
