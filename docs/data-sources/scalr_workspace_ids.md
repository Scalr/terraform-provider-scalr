
# Data Source `scalr_workspace_ids` 

Retrieves a map of workspace IDs based on the names or/and tags provided. Wildcards are accepted.

## Example Usage

```hcl
data "scalr_workspace_ids" "app-frontend" {
  names          = ["app-frontend-prod", "app-frontend-dev1", "app-frontend-staging"]
  environment_id = "env-xxxxxxxxxxx"
}

data "scalr_workspace_ids" "all" {
  names          = ["*"]
  environment_id = "env-xxxxxxxxxxx"
}

data "scalr_workspace_ids" "partial-match" {
  names          = ["dev-"]
  exact_match    = false
  environment_id = "env-xxxxxxxxxxx"
}

data "scalr_workspace_ids" "tagged" {
  tag_ids         = ["tag-xxxxxxxxxxx", "tag-yyyyyyyyyyy"]
  environment_id  = "env-xxxxxxxxxxx"
}
```

## Argument Reference

* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.
* `names` - (Optional) A list of names to search for. If a name does not exist, it will not throw an error, it will just not exist in the returned output. Use `["*"]` to select all workspaces.
* `tag_ids` - (Optional) List of tag IDs associated with the workspace.
* `exact_match` - (Optional) If `true`, performs exact match on workspace names, otherwise will match names that contain given values. Defaults to `true`.

## Attribute Reference

All arguments plus:

* `ids` - A map of workspace names and their opaque IDs, in the format `env_id/name`.
