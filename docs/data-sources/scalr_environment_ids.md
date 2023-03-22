
# Data Source `scalr_environment_ids` 

Retrieves a map of environment IDs based on the names or/and tags provided. Wildcards are accepted.

## Example Usage

```hcl
data "scalr_environment_ids" "exact-names" {
  names       = ["production", "development"]
}

data "scalr_environment_ids" "all" {
  names       = ["*"]
}

data "scalr_environment_ids" "partial-match" {
  names       = ["dev-"]
  exact_match = false
}

data "scalr_environment_ids" "tagged" {
  tag_ids     = ["tag-xxxxxxxxxxx", "tag-yyyyyyyyyyy"]
}
```

## Argument Reference

* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.
* `names` - (Optional) A list of names to search for. If a name does not exist, it will not throw an error, it will just not exist in the returned output. Use `["*"]` to select all environments.
* `tag_ids` - (Optional) List of tag IDs associated with the environment.
* `exact_match` - (Optional) If `true`, performs exact match on environment names, otherwise will match names that contain given values. Defaults to `true`.

## Attribute Reference

All arguments plus:

* `ids` - A map of environment names and their opaque IDs, in the format `acc_id/name`.
