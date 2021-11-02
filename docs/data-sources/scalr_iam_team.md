---
layout: "scalr"
page_title: "Scalr: scalr_iam_team"
sidebar_current: "docs-datasource-scalr-iam-team-x"
description: |-
Get information on a team.
---

# scalr_iam_team Data Source

This data source is used to retrieve details of a team by name and account_id.

## Example Usage

```hcl
data "scalr_iam_team" "example" {
  name        = "dev"
  account_id  = "acc-xxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) Name of the team.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - The ID of the team.
* `description` - Verbose description of the team.
* `identity_provider_id` - ID of the identity provider, in the format `idp-<RANDOM STRING>`.
* `users` - Array of user ids that belong to this team.
