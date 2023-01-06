
# Data Source `scalr_service_account` 

Retrieves information about a service account.

## Example Usage

```hcl
data "scalr_service_account" "example" {
  email      = "name@account.scalr.io"
  account_id = "acc-xxxxxxxxx"
}
```

## Arguments

* `email` - (Required) The email of the service account.
* `account_id` - (Optional) The ID of the Scalr account, in the format `acc-<RANDOM STRING>`

## Attributes

All arguments plus:

* `id` - The identifier of the service account in the format `sa-<RANDOM STRING>`.
* `created_by` - Details of the user that created the service account.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.
