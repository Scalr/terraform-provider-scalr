
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

* `id` - (Optional) The identifier of the service account in the format `sa-<RANDOM STRING>`.
* `email` - (Optional) The email of the service account.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.

Arguments `id` and `email` are both optional, specify at least one of them to obtain `scalr_service_account`.

## Attributes

All arguments plus:

* `name` - Name of the service account.
* `description` - Description of the service account.
* `status` - The status of the service account.
* `created_by` - Details of the user that created the service account.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.
