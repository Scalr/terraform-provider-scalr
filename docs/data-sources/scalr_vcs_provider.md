
# Data Source `scalr_vcs_provider` 

Retrieves the details of a VCS provider.

## Example Usage

```hcl
data "scalr_vcs_provider" "example" {
  id = "vcs-xxxxxxx"
  account_id = "acc-xxxxxxx"
}
```

```hcl
data "scalr_vcs_provider" "example" {
  name = "example"
  account_id = "acc-xxx"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) Identifier of the VCS provider.
* `name` - (Optional) Name of the VCS provider.
* `vcs_type` - (Optional) Type of the VCS provider. For example, `github`.
* `environment_id` - (Optional) ID of the environment the VCS provider has to be linked to, in the format `env-<RANDOM STRING>`.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.
* `agent_pool_id` - (Optional) The id of the agent pool to connect Scalr to self-hosted VCS provider, in the format `apool-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `url` - The URL to the VCS provider installation.
* `environments` - List of the identifiers of environments the VCS provider is linked to.
