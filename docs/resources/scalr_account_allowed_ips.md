---
layout: "scalr"
page_title: "Scalr: scalr_account_allowed_ips"
sidebar_current: "docs-resource-scalr-account-allowed-ips"
description: |-
  Manages allowed ips for account.
---

# scalr_account_allowed_ips Resource

Manage the state of allowed ips for an account in Scalr. Create, update and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_account_allowed_ips" "default" {
  account_id  = "acc-xxxxxxxx"
  allowed_ips = ["127.0.0.1", "192.168.0.0/24"]
}
```

## Argument Reference

* `account_id` - (Required) ID of the account.
* `allowed_ips` - (Required) The list of allowed IPs or CIDRs.

## Attribute Reference

All arguments plus:

* `id` - The ID of the account.

## Import

To import allowed ips for an account use account ID as the import ID. For example:
```shell
terraform import scalr_account_allowed_ips.default acc-xxxxxxxxx
```
