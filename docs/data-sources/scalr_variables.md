
# Data Source `scalr_variables`

Retrieves the list of variables by the given filters.

## Example Usage

```hcl
data "scalr_variables" "vars" {
  keys = ["key1", "key2", "key3"]
  category = "terraform" # or shell
  account_id = "acc-tgobtsrgo3rlks8"
  envrironment_ids = ["env-sv0425034857d22", "null"]
  workspace_ids = ["ws-tlbp7litrs55vgg", "null"]
}
```

## Argument Reference

* `keys` - (Optional) A list of keys to be used in the query used in a Scalr variable name filter.
* `category` - (Optional) The category of a Scalr variable.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`
* `envrironment_ids` - (Optional) A list of identifiers of the Scalr environments, in the format `env-<RANDOM STRING>`. Used to shrink the variable's scope in case the variable name exists in multiple environments.
* `workspace_ids` - (Optional) A list of identifiers of the Scalr workspace, in the format `ws-<RANDOM STRING>`. Used to shrink the variable's scope in case the variable name exists on multiple workspaces.


## Attribute Reference


* `variables` - The list of Scalr variables with all attributes.

The `variables` block item contains:

* `id` - ID of the variable.
* `key` - Key of the variable.
* `value` - Variable value if it is not sensitive.
* `category` - Indicates if this is a Terraform or shell variable.
* `description` - Variable verbose description.
* `hcl` - If the variable is configured as a string of HCL code.
* `sensitive` - If the variable is configured as sensitive.
* `final` - If the variable is configured as final. Indicates whether the variable can be overridden on a lower scope down the Scalr organizational model.
* `workspace_id` - The workspace that owns the variable, specified as an ID, in the format `ws-<RANDOM STRING>`.
* `environment_id` - The environment that owns the variable, specified as an ID, in the format `env-<RANDOM STRING>`.
* `account_id` - The account that owns the variable, specified as an ID, in the format `acc-<RANDOM STRING>`.
