---
layout: "scalr"
page_title: "Scalr: scalr_environment"
sidebar_current: "docs-datasource-scalr-environment-x"
description: |-
  Get information on an environment.
---

# scalr_environment

This data source is used to retrieve details of a an environment.

## Example Usage

```hcl
data "scalr_environment" "test" {
  id = "env-xxxxxxxxxx"
}`
```

## Arguments

* `id` - (Required) The environment ID, in the format `env-<RANDOM STRING>`.

## Attributes

All arguments plus:

* `name` - Name of the environment.
* `created_by` - Details of the user that created the environment.
* `cost_estimation_enabled` - Whether cost estimation for the environment  enabled (true/false).
* `status` - Shows status of the environment. 
* `account_id` - ID of the environment account, in the format `acc-<RANDOM STRING>`
* `cloud_credentials` - List of the environment cloud-credentials IDs, in the format `cred-<RANDOM STRING>`.
* `policy_groups` - List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.