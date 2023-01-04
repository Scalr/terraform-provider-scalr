
# Data Source `scalr_current_account` 

Retrieves the details of current account when using Scalr remote backend.

## Example Usage

```hcl
data "scalr_current_account" "account" {}
```

## Argument Reference

No arguments are required. The data source returns details of the current account
based on the `SCALR_ACCOUNT_ID` environment variable that is automatically exported in the Scalr remoted backend.

## Attribute Reference

* `id` - The identifier of the account.
* `name` - The name of the account.
