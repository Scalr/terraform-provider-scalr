---
title: "scalr_policy_group_linkage"
categorySlug: "scalr-terraform-provider"
slug: "provider_resource_scalr_policy_group_linkage"
parentDocSlug: "provider_resources"
hidden: false
order: 10
---
## Resource Overview

Manage policy group to environment linking in Scalr. Create, update and destroy.

## Example Usage

```terraform
resource "scalr_policy_group_linkage" "example" {
  policy_group_id = "pgrp-xxxxxxxxxx"
  environment_id  = "env-xxxxxxxxxx"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (String) ID of the environment, in the format `env-<RANDOM STRING>`.
- `policy_group_id` (String) ID of the policy group, in the format `pgrp-<RANDOM STRING>`.

### Read-Only

- `id` (String) The ID of the policy group linkage. It is a combination of the policy group and environment IDs in the format `pgrp-xxxxxxxxxxxxxxx/env-yyyyyyyyyyyyyyy`

## Import

Import is supported using the following syntax:

```shell
terraform import scalr_policy_group_linkage.example pgrp-xxxxxxxxxx/env-xxxxxxxxxx
```