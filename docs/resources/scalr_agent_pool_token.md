
# Resource `scalr_agent_pool_token` 

Manage the state of agent pool's tokens in Scalr. Create, update and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_agent_pool_token" "default" {
  description   = "Some description"
  agent_pool_id = "apool-xxxxxxx"
}
```

## Argument Reference

* `description` - (Required) Description of the token.
* `agent_pool_id` - (Required) ID of the agent pool.

## Attribute Reference

All arguments plus:

* `id` - The ID of the token.
* `token` - The token of the agent pool.
