---
layout: "scalr"
page_title: "Scalr: scalr_iam_team"
sidebar_current: "docs-resource-scalr-iam-team"
description: |-
Manages teams.
---

# scalr_iam_team Resource

Manage the Scalr IAM teams. Create, update and destroy.

## Example Usage

```hcl
resource "scalr_iam_team" "example" {
  name        = "dev"
  description = "Developers"
  account_id  = "acc-xxxxxxxx"

  users = [ "user-xxxxxxxx", "user-yyyyyyyy" ]
}
```

## Argument Reference

* `name` - (Required) Name of the team.
* `description` - (Optional) Verbose description of the team.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.
* `identity_provider_id` - (Optional) ID of the identity provider, in the format `idp-<RANDOM STRING>`. This is required when `account_id` is not set.
* `users` - (Optional) Array of user ids to add to this team.

## Attribute Reference

All arguments plus:

* `id` - The ID of the team.

## Import

To import teams use team ID as the import ID. For example:
```shell
terraform import scalr_iam_team.example team-tntulnted6oom28
```
