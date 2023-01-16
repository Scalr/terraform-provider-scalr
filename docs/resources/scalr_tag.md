
# Resource `scalr_tag`

Manages the state of tags in Scalr.

## Example Usage

Basic usage:

```hcl
resource "scalr_tag" "example" {
  name       = "tag-name"
  account_id = "acc-xxxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) Name of the tag.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.

## Attributes

All arguments plus:

* `id` - The identifier of the tag in the format `tag-<RANDOM STRING>`.
