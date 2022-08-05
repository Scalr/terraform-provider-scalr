
# Data Source `scalr_iam_user` 

Retrieves the details of a Scalr user by the email.

## Example Usage

```hcl
data "scalr_iam_user" "example" {
  email = "user@test.com"
}
```

## Argument Reference

* `email` - (Required) The email of a user.

## Attribute Reference

All arguments plus:

* `id` - An identifier of the user.
* `status` - A system status of the user.
* `username` - A username of the user.
* `full_name` - A full name of the user.
* `identity_providers` - A list of the identity providers the user belongs to.
* `teams` - A list of the team identifiers the user belongs to.
