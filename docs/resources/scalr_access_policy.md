---
layout: "scalr"
page_title: "Scalr: scalr_access_policy"
sidebar_current: "docs-resource-scalr-access-policy"
description: |-
  Manages access policies.
---

# scalr_access_policy Resource

Manage the Scalr IAM access policies. Create, update and destroy

## Example Usage

Basic usage:

```hcl
resource "scalr_role" "reader" {
  name         = "Reader"
  account_id = "acc-xxxxxxxx"
  description = "Read access to all resources."

  permissions = [
    "*:read",
  ]
}

resource "scalr_access_policy" "team_read_all_on_acc_scope" {
  subject {
    type = "team"
    id = "team-xxxxxxx"
  }
  scope {
    type = "account"
    id = "acc-xxxxxxx"
  }

  role_ids = [
    scalr_role.reader.id
  ]
}
```

## Argument Reference

* `scope` - (Required) Defines the scope where access policy is applied.
* `subject` - (Required) Defines the subject of the access policy.
* `role_ids` - (Required) The list of the role IDs.


## Attribute Reference

All arguments plus:

* `id` - The access policy ID.

The `scope` block contains:

* `type` - The scope identity type, is one of `account`, `environment`, or `workspace`.
* `id` - The scope ID, `acc-<RANDOM STRING>` for account, `env-<RANDOM STRING>` for environment, `ws-<RANDOM STRING>` for workspace.

The `subject` block contains:

* `type` - The subject type, is one of `user`, `team`, or `service_account`.
* `id` - The subject ID, `user-<RANDOM STRING>` for user, `team-<RANDOM STRING>` for team, `sa-<RANDOM STRING>` for service account.## Import

To import an access policy use access policy ID as the import ID. For example:
```shell
terraform import scalr_access_policy.example ap-te2cteuismsqocd
```
