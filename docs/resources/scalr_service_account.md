
# Resource `scalr_service_account`

Manages the state of service accounts in Scalr.

## Example Usage

Basic usage:

```hcl
resource "scalr_service_account" "example" {
  name        = "sa-name"
  description = "Lorem ipsum"
  status      = "Active"
  account_id  = "acc-xxxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) Name of the service account.
* `description` - (Optional) Description of the service account.
* `status` - (Optional) The status of the service account. Valid values are `Active` and `Inactive`.
Defaults to `Active`.
* `account_id` - (Optional) ID of the environment account, in the format `acc-<RANDOM STRING>`

## Attributes

All arguments plus:

* `id` - The identifier of the service account in the format `sa-<RANDOM STRING>`.
* `created_by` - Details of the user that created the service account.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.
