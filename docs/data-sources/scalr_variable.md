
# Data Source `scalr_variable`

Retrieves the details of a variable.

## Example Usage

```hcl
data "scalr_variable" "test_var" {
  key = "key"
  category = "terraform"
  account_id = "acc-tgobtsrgo3rlks8"
  environment_id = "env-sv0425034857d22"
  workspace_id = "ws-tlbp7litrs55vgg"
}
```

## Argument Reference

* `key` - (Required) The name of a Scalr variable.
* `category` - (Optional) The category of a Scalr variable.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`
* `environment_id` - (Optional) The identifier of the Scalr environment, in the format `env-<RANDOM STRING>`. Used to shrink the scope of the variable in case the variable name exists in multiple environments.
* `workspace_id` - (Optional) The identifier of the Scalr workspace, in the format `ws-<RANDOM STRING>`. Used to shrink the scope of the variable in case the variable name exists on multiple workspaces.


## Attribute Reference

All arguments plus:

* `id` - ID of the variable.
* `value` - Variable value.
* `description` - Variable verbose description, defaults to empty string.
* `hcl` - If the variable is configured as a string of HCL code.
* `sensitive` - If the variable is configured as sensitive.
* `final` - If the variable is configured as final. Indicates whether the variable can be overridden on a lower scope down the Scalr organizational model.
