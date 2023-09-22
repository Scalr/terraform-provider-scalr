---
title: "scalr_access_policy"
category: "6380b9efad50240652eec1fc"
slug: "provider_datasource_scalr_access_policy"
parentDocSlug: "provider_datasources"
hidden: false
order: 1
---
## Data Source Overview

This data source is used to retrieve details of a single access policy by id.

## Example Usage

```terraform
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

<!-- Manually filling the schema here because of https://github.com/hashicorp/terraform-plugin-docs/issues/28 -->
## Schema

### Required

- `id` (String) The access policy ID.

### Read-Only

- `is_system` (Boolean)
- `role_ids` (List of String) The list of the role IDs.
- `scope` (List of Object) Defines the scope where access policy is applied. (see [below for nested schema](#nestedatt--scope))
- `subject` (List of Object) Defines the subject of the access policy. (see [below for nested schema](#nestedatt--subject))

<a id="nestedatt--scope"></a>
### Nested Schema for `scope`

Read-Only:

- `id` (String) The scope ID, `acc-<RANDOM STRING>` for account, `env-<RANDOM STRING>` for environment, `ws-<RANDOM STRING>` for workspace.
- `type` (String) The scope identity type, is one of `account`, `environment`, or `workspace`.


<a id="nestedatt--subject"></a>
### Nested Schema for `subject`

Read-Only:

- `id` (String) The subject ID, `user-<RANDOM STRING>` for user, `team-<RANDOM STRING>` for team, `sa-<RANDOM STRING>` for service account.
- `type` (String) The subject type, is one of `user`, `team`, or `service_account`.
