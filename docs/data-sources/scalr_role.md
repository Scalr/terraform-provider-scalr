
# Data Source `scalr_role` 

This data source is used to retrieve details of a single role.

## Example Usage

To retrieve a custom role, an account id and role id (or name) are required, for example: 

```hcl
data "scalr_role" "example" {
  id         = "role-xxxxxxx"
  account_id = "acc-xxxxxxx"
}
```

```hcl
data "scalr_role" "example" {
  name       = "WorkspaceAdmin"
  account_id = "acc-xxxxxxx"
}
```

To retrieve system-managed roles an account id has to be omitted, for example:

```hcl
data "scalr_role" "example" {
  name = "user"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) ID of the role.
* `name` - (Optional) Name of the role.
* `account_id` - (Optional) ID of the account.

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_role`.

## Attribute Reference

All arguments plus:

* `permissions` - Array of permission names.
* `is_system` - Boolean indicates if the role can be edited.
* `description` - Verbose description of the role.
