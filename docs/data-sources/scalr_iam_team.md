---
layout: "scalr"
page_title: "Scalr: scalr_iam_team"
sidebar_current: "docs-datasource-scalr-iam-team-x"
description: |-
Get information on a team.
---

# scalr_iam_team Data Source

Retrieves the details of a team by the name and account_id.

## Example Usage

```hcl
data "scalr_iam_team" "example" {
  name        = "dev"
  account_id  = "acc-xxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) Name of the team.
* `account_id` - (Optional) The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - An identifier of the team.
* `description` - A verbose description of the team.
* `identity_provider_id` - An identifier of an identity provider team is linked to, in the format `idp-<RANDOM STRING>`.
* `users` - The list of the user identifiers that belong to the team.
