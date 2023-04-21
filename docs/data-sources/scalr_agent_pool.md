
# Data Source `scalr_agent_pool` 

Retrieves the details of an agent pool.

## Example Usage

Basic usage:

```hcl
data "scalr_agent_pool" "default" {
  id         = "apool-xxxxxxx"
  account_id = "acc-xxxxxxx"
}
```

```hcl
data "scalr_agent_pool" "default" {
  name       = "default-pool"
  account_id = "acc-xxxxxxx"
}
```

## Argument Reference

* `id` - (Optional) ID of the agent pool.
* `name` - (Optional) A name of the agent pool.
* `account_id` - (Optional) An identifier of the Scalr account.
* `environment_id` - (Optional) An identifier of the Scalr environment.

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_agent_pool`.

## Attribute Reference

All arguments plus:

* `workspace_ids` - The list of IDs of linked workspaces.