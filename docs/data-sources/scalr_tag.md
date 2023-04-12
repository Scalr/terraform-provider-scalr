
# Data Source `scalr_tag` 

Retrieves information about a tag.

## Example Usage

```hcl
data "scalr_tag" "example" {
  id         = "tag-xxxxxxx"
  account_id = "acc-xxxxxxx"
}
```

```hcl
data "scalr_tag" "example" {
  name       = "tag-name"
  account_id = "acc-xxxxxxx"
}
```

## Arguments

* `id` - (Optional) The identifier of the tag in the format `tag-<RANDOM STRING>`.
* `name` - (Optional) The name of the tag.
* `account_id` - (Optional) The ID of the Scalr account, in the format `acc-<RANDOM STRING>`.

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_tag`.

## Attributes

Attributes are the same as arguments.
