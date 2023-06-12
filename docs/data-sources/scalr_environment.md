
# Data Source `scalr_environment`

Retrieves the details of a Scalr environment.

## Example Usage

```hcl
data "scalr_environment" "test" {
  id         = "env-xxxxxxx"
  account_id = "acc-xxxxxxx"
}
```

```hcl
data "scalr_environment" "test" {
  name       = "environment-name"
  account_id = "acc-xxxxxxx"
}
```

## Arguments

* `id` - (Optional) The environment ID, in the format `env-<RANDOM STRING>`.
* `name` - (Optional) Name of the environment.
* `account_id` - (Optional) ID of the environment account, in the format `acc-<RANDOM STRING>`

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_environment`.

## Attributes

All arguments plus:

* `created_by` - Details of the user that created the environment.
* `cost_estimation_enabled` - Boolean indicates if cost estimation is enabled for the environment.
* `status` - The status of an environment.
* `policy_groups` - List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.
* `tag_ids` - List of tag IDs associated with the environment.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.
