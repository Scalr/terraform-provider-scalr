
# Data Source `scalr_tag` 

Retrieves information about the tag.

## Example Usage

```hcl
data "scalr_tag" "example" {
  name = "tag-name"
  account_id = "acc-xxxxxxxxx"
}
```

## Arguments

* `name` - (Required) A name of the tag.
* `account_id` - (Required) An ID of the Sclar account, in the format `acc-<RANDOM STRING>`

## Attributes

All arguments plus:

* `id` - The identifier of the tag in the format `tag-<RANDOM STRING>`.
