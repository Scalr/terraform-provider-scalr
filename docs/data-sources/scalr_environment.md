---
layout: "scalr"
page_title: "Scalr: scalr_environment"
sidebar_current: "docs-datasource-scalr-environment-x"
description: |-
  Get information on an environment.
---

# scalr_environment Data Source

This data source is used to retrieve details of a an environment.

## Example Usage

```hcl
data "scalr_environment" "test" {
  id = "env-xxxxxxxxxx" # optional, conflicts with filter by name
  account_id = "acc-xxxxxxxx" # mandatory if user has access to few accounts and environment name is not unique
  name = "environment-name"  # optional, conflicts with filter by id
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
* `status` - Shows status of the environment. 
* `cloud_credentials` - List of the environment cloud-credentials IDs, in the format `cred-<RANDOM STRING>`.
* `policy_groups` - List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.
