
# Data Source `scalr_environment`

Retrieves the details of a Scalr environment.

## Example Usage

```hcl
data "scalr_environment" "test" {
  id = "env-xxxxxxxxxx" # optional, can only use id or name for the environment filter, if both are used there will be a conflict.
  account_id = "acc-xxxxxxxx" # mandatory if a user has access to a few accounts and the environment name is not unique
  name = "environment-name"  # optional, can only use id or name for the environment filter, if both are used there will be a conflict.
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
* `cloud_credentials` - List of the environment cloud-credentials IDs, in the format `cred-<RANDOM STRING>`.
* `policy_groups` - List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.
