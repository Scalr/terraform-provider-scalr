---
layout: "scalr"
page_title: "Scalr: scalr_iam_user"
sidebar_current: "docs-datasource-scalr-iam-user-x"
description: |-
Get information on a user.
---

# scalr_iam_user Data Source

This data source is used to retrieve details of a user by email.

## Example Usage

```hcl
data "scalr_iam_user" "example" {
  email = "user@test.com"
}
```

## Argument Reference

* `email` - (Required) Email of the user.

## Attribute Reference

All arguments plus:

* `id` - The ID of the user.
* `status` - Shows status of the user.
* `username` - Users username.
* `full_name` - Users full name.
* `identity_providers` - List of identity provider ids this user belongs to.
* `teams` - List of team ids this user belongs to.
