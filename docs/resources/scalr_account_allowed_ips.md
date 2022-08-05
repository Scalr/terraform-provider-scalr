
# Resource `scalr_account_allowed_ips` 

Manages the list of allowed IPs for an account in Scalr. Create, update and destroy.

## Example Usage

Basic usage:

```javascript
resource "scalr_account_allowed_ips" "default" {
  account_id  = "acc-xxxxxxxx"
  allowed_ips = ["127.0.0.1", "192.168.0.0/24"]
}
```

## Argument Reference

* `account_id` -  (Required) ID of the account.
* `allowed_ips` - (Required) The list of allowed IPs or CIDRs. 
                  **Warning**: if you don't specify the current IP address, you may lose access to the account. 
                  To restore it the account owner has to raise a [support ticket](https://suport.scalr.com)

## Attribute Reference

All arguments plus:

* `id` - The ID of the account.

## Import

To import allowed ips for an account use account ID as the import ID. For example:

```shell
terraform import scalr_account_allowed_ips.default acc-xxxxxxxxx
```
