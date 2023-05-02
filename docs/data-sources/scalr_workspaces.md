
# Data Source `scalr_workspaces` 

Retrieves a list of workspace ids by name or tags.

## Example Usage

```hcl
data "scalr_workspaces" "exact-names" {
  name = "in:production,development"
}

data "scalr_workspaces" "app-frontend" {
  name           = "like:app-frontend-"
  environment_id = "env-xxxxxxxxxxx"
}

data "scalr_workspaces" "tagged" {
  tag_ids = ["tag-xxxxxxxxxxx", "tag-yyyyyyyyyyy"]
}

data "scalr_workspaces" "all" {
  environment_id = "env-xxxxxxxxxxx"
}
```

## Argument Reference

* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.
* `environment_id` - (Optional) ID of the environment, in the format `env-<RANDOM STRING>`.
* `name` - (Optional) The query used in a Scalr workspace name filter.
* `tag_ids` - (Optional) List of tag IDs associated with the workspace.

## Attribute Reference

All arguments plus:

* `ids` - The list of workspace IDs, in the format [`ws-xxxxxxxxxxx`, `ws-yyyyyyyyy`].
