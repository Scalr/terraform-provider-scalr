
# Resource `scalr_service_account_token` 

Manage the state of service account's tokens in Scalr. Create, update and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_service_account_token" "default" {
  service_account_id = "sa-xxxxxxx"
  description        = "Some description"
}
```

## Argument Reference

* `service_account_id` - (Required) ID of the service account.
* `description` - (Optional) Description of the token.

## Attribute Reference

All arguments plus:

* `id` - The ID of the token.
* `token` - (Sensitive) The token of the service account.
