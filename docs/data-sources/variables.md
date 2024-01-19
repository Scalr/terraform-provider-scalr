---
title: "scalr_variables"
categorySlug: "scalr-terraform-provider"
slug: "provider_datasource_scalr_variables"
parentDocSlug: "provider_datasources"
hidden: false
order: 18
---
## Data Source: scalr_variables

Retrieves the list of variables by the given filters.

## Example Usage

```terraform
data "scalr_variables" "example" {
  keys             = ["key1", "key2", "key3"]
  category         = "terraform" # or shell
  account_id       = "acc-xxxxxxxxxx"
  envrironment_ids = ["env-xxxxxxxxxx", "null"]
  workspace_ids    = ["ws-xxxxxxxxxx", "null"]
}
```

<!-- Manually filling the schema here because of https://github.com/hashicorp/terraform-plugin-docs/issues/28 -->
## Schema

### Optional

- `account_id` (String) ID of the account, in the format `acc-<RANDOM STRING>`.
- `category` (String) The category of a Scalr variable.
- `environment_ids` (Set of String) A list of identifiers of the Scalr environments, in the format `env-<RANDOM STRING>`. Used to shrink the variable's scope in case the variable name exists in multiple environments.
- `keys` (Set of String) A list of keys to be used in the query used in a Scalr variable name filter.
- `workspace_ids` (Set of String) A list of identifiers of the Scalr workspace, in the format `ws-<RANDOM STRING>`. Used to shrink the variable's scope in case the variable name exists on multiple workspaces.

### Read-Only

- `id` (String) The ID of this resource.
- `variables` (Set of Object) The list of Scalr variables with all attributes. (see [below for nested schema](#nestedatt--variables))

<a id="nestedatt--variables"></a>
### Nested Schema for `variables`

Read-Only:

- `account_id` (String) The account that owns the variable, specified as an ID, in the format `acc-<RANDOM STRING>`.
- `category` (String) Indicates if this is a Terraform or shell variable.
- `description` (String) Variable verbose description.
- `environment_id` (String) The environment that owns the variable, specified as an ID, in the format `env-<RANDOM STRING>`.
- `final` (Boolean) If the variable is configured as final. Indicates whether the variable can be overridden on a lower scope down the Scalr organizational model.
- `hcl` (Boolean) If the variable is configured as a string of HCL code.
- `id` (String) ID of the variable.
- `key` (String) Key of the variable.
- `sensitive` (Boolean) If the variable is configured as sensitive.
- `value` (String) Variable value if it is not sensitive.
- `workspace_id` (String) The workspace that owns the variable, specified as an ID, in the format `ws-<RANDOM STRING>`.
