---
layout: "scalr"
page_title: "Scalr: scalr_access_policy"
sidebar_current: "docs-datasource-scalr-access-policy-x"
description: |-
  Get information on a IAM access policy.
---

# scalr_access_policy Data Source

This data source is used to retrieve details of a single access policy by id.

## Example Usage

```hcl
data "scalr_access_policy" "example" {
  id = "ap-xxxxxxxxx"
}

output "scope_id" {
  value = data.scalr_access_policy.example.scope[0].id
}

output "subject_id" {
  value = data.scalr_access_policy.example.subject[0].id
}
```

## Argument Reference

The following arguments are supported:

* `id` - The access policy ID.

## Attribute Reference

All arguments plus:

* `scope` - Defines the scope where access policy is applied.
* `subject` - Defines the subject of the access policy.
* `role_ids` - The list of the role IDs.

The `scope` block contains:

* `type` - The scope identity type, is one of `account`, `environment`, or `workspace`.
* `id` - The scope ID, `acc-<RANDOM STRING>` for account, `env-<RANDOM STRING>` for environment, `ws-<RANDOM STRING>` for workspace.

The `subject` block contains:

* `type` - The subject type, is one of `user`, `team`, or `service_account`.
* `id` - The subject ID, `user-<RANDOM STRING>` for user, `team-<RANDOM STRING>` for team, `sa-<RANDOM STRING>` for service account.
