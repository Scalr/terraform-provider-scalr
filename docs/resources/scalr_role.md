
# Resource `scalr_role`

Manage the Scalr IAM roles. Create, update and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_role" "writer" {
  name         = "Writer"
  account_id = "acc-xxxxxxxx"
  description = "Write access to all resources."

  permissions = [
    "*:update",
    "*:delete",
    "*:create"
  ]
}
```

## Argument Reference

* `name` - (Required) Name of the role.
* `account_id` - (Optional) ID of the account.
* `permissions` - (Required) Array of permission names.
* `description` - (Optional) Verbose description of the role.

## Attribute Reference

All arguments plus:

* `id` - The ID of the role.
* `is_system` - Boolean indicates if the role can be edited. System roles are maintained by Scalr and cannot be changed.

## Import

To import roles use role ID as the import ID. For example:
```shell
terraform import scalr_role.example role-te2cteuismsqocg
```
