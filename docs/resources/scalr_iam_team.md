---
layout: "scalr"
page_title: "Scalr: scalr_iam_team"
sidebar_current: "docs-resource-scalr-iam-team"
description: |-
Manages teams.
---

# scalr_iam_team Resource

Manages the Scalr IAM teams: performs create, update and destroy actions.

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

* `name` - (Required) A name of the team.
* `description` - (Optional) A verbose description of the team.
* `account_id` - (Optional) An identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.
* `identity_provider_id` - (Optional) An identifier of the login identity provider, in the format `idp-<RANDOM STRING>`. This is required when `account_id` is not specified.
* `users` - (Optional) A list of the user identifiers to add to the team.

## Attribute Reference

All arguments plus:

* `id` - The ID of the team.

## Import

To import teams use team ID as the import ID. For example:

```shell
terraform import scalr_iam_team.example team-tntulnted6oom28
```
