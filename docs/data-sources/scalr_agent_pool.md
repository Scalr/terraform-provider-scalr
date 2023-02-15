
# Data Source `scalr_agent_pool` 

Retrieves the details of an agent pool.

## Example Usage

Basic usage:

```hcl
data "scalr_agent_pool" "default" {
  name       = "default-pool"
  account_id = "acc-xxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) A name of the agent pool.
* `account_id` - (Optional) An identifier of the Scalr account.
* `environment_id` - (Optional) An identifier of the Scalr environment.

## Attribute Reference

All arguments plus:

* `id` - The ID of the agent pool.
* `workspace_ids` - The list of IDs of linked workspaces.