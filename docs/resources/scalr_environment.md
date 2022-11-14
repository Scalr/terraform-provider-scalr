
# Resource `scalr_environment`

Manage the state of environments in Scalr. Creates, updates and destroys.

## Example Usage

Basic usage:

```hcl
resource "scalr_environment" "test" {
  name       = "test-env"
  account_id = "acc-<id>"
  cost_estimation_enabled = true
  cloud_credentials = ["cred-xxxxx", "cred-yyyyy"]
  policy_groups = ["pgrp-xxxxx", "pgrp-yyyyy"]
  default_provider_configurations = ["pcfg-xxxxx", "pcfg-yyyyy"]
}
```

## Argument Reference

* `name` - (Required) Name of the environment.
* `account_id` - (Required) ID of the environment account, in the format `acc-<RANDOM STRING>`
* `cost_estimation_enabled` - (Optional) Set (true/false) to enable/disable cost estimation for the environment. Default `true`.
* `cloud_credentials` - Deprecated and will be removed at 2023-01-12. Use `default_provider_configurations` instead.
* `policy_groups` - (Optional) List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.
* `default_provider_configurations` - (Optional) List of IDs of provider configurations, used in the environment workspaces by default.
* `tag_ids` - (Optional) List of tag IDs associated with the environment.

## Attributes

All arguments plus:

* `id` - The environment ID, in the format `env-<RANDOM STRING>`.
* `created_by` - Details of the user that created the environment.
* `status` - The status of the environment. 

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.
