
# Data Source `scalr_iam_teams` 

Retrieves a list of a team ids by the name.

## Example Usage

```hcl
data "scalr_iam_teams" "example" {
  name        = "in:dev,stage"
  account_id  = "acc-xxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) The query used in a Scalr iam teams name filter.
* `account_id` - (Optional) The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `ids` - The list of iam team IDs, in the format [`team-xxxxxxxxxxx`, `team-yyyyyyyyy`].
