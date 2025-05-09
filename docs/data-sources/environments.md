---
title: "scalr_environments"
categorySlug: "scalr-terraform-provider"
slug: "provider_datasource_scalr_environments"
parentDocSlug: "provider_datasources"
hidden: false
order: 7
---
## Data Source: scalr_environments

Retrieves a list of environment ids by name or tags.

## Example Usage

```terraform
data "scalr_environments" "exact-names" {
  name = "in:production,development"
}

data "scalr_environments" "app-frontend" {
  name = "like:app-frontend-"
}

data "scalr_environments" "tagged" {
  tag_ids = ["tag-xxxxxxxxxx", "tag-yyyyyyyyyy"]
}

data "scalr_environments" "all" {
  account_id = "acc-xxxxxxxxxx"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `account_id` (String) The ID of the Scalr account, in the format `acc-<RANDOM STRING>`.
- `name` (String) The query used in a Scalr environment name filter.
- `tag_ids` (Set of String) List of tag IDs associated with the environment.

### Read-Only

- `id` (String) The identifier of this data source.
- `ids` (Set of String) The list of environment IDs, in the format [`env-xxxxxxxxxxx`, `env-yyyyyyyyy`].
