
# Data Source `scalr_iam_user` 

Retrieves the details of a Scalr user.

## Example Usage

```hcl
data "scalr_iam_user" "example" {
  id = "user-xxxxxxx"
}
```

```hcl
data "scalr_iam_user" "example" {
  email = "user@test.com"
}
```

## Argument Reference

* `id` - (Optional) An identifier of a user.
* `email` - (Optional) An email of a user.

Arguments `id` and `email` are both optional, specify at least one of them to obtain `scalr_iam_user`.

## Attribute Reference

All arguments plus:

* `status` - A system status of the user.
* `username` - A username of the user.
* `full_name` - A full name of the user.
* `identity_providers` - A list of the identity providers the user belongs to.
* `teams` - A list of the team identifiers the user belongs to.
