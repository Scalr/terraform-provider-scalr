---
layout: "scalr"
page_title: "Scalr: scalr_environment"
sidebar_current: "docs-resource-scalr-environment"
description: |-
  Manages environments.
---

# scalr_environment

Manage the state of environments in Scalr. Creates, updates and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_environment" "test" {
  name       = "test-env"
  account_id = "acc-<id>"
  cost_estimation_enabled = true
  cloud-credentials = ["cred-xxxxx", "cred-yyyyy"]
  policy_groups = ["pgrp-xxxxx", "pgrp-yyyyy"]

}
```

## Argument Reference

* `name` - Name of the environment.
* `cost_estimation_enabled` - Whether cost estimation for the environment  enabled (true/false).
* `account_id` - ID of the environment account, in the format `acc-<RANDOM STRING>`
* `cloud_credentials` - List of the environment cloud-credentials IDs, in the format `cred-<RANDOM STRING>`.
* `policy_groups` - List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.

## Attributes

All arguments plus:

* `id` - The environment ID, in the format `env-<RANDOM STRING>`.
* `created_by` - Details of the user that created the environment.
* `status` - Shows status of the environment. 
