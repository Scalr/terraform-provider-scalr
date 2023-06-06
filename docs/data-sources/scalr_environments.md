
# Data Source `scalr_environments` 

Retrieves a list of environment ids by name or tags.

## Example Usage

```hcl
data "scalr_environments" "exact-names" {
  name = "in:production,development"
}

data "scalr_environments" "app-frontend" {
  name = "like:app-frontend-"
}

data "scalr_environments" "tagged" {
  tag_ids = ["tag-xxxxxxxxxxx", "tag-yyyyyyyyyyy"]
}

data "scalr_environments" "all" {
  account_id = "acc-xxxxxxxxxxx"
}
```

## Argument Reference

* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.
* `name` - (Optional) The query used in a Scalr environment name filter.
* `tag_ids` - (Optional) List of tag IDs associated with the environment.

## Attribute Reference

All arguments plus:

* `ids` - The list of environment IDs, in the format [`env-xxxxxxxxxxx`, `env-yyyyyyyyy`].
