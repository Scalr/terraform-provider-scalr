
# Data Source `scalr_tag` 

Retrieves information about a tag.

## Example Usage

```hcl
data "scalr_tag" "example" {
  name = "tag-name"
  account_id = "acc-xxxxxxxxx"
}
```

## Arguments

* `name` - (Required) The name of the tag.
* `account_id` - (Optional) The ID of the Scalr account, in the format `acc-<RANDOM STRING>`

## Attributes

All arguments plus:

* `id` - The identifier of the tag in the format `tag-<RANDOM STRING>`.
