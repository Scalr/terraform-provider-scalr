---
title: "scalr_hook"
categorySlug: "scalr-terraform-provider"
slug: "provider_datasource_scalr_hook"
parentDocSlug: "provider_datasources"
hidden: false
order: 9
---
## Data Source: scalr_hook

Retrieves information about a hook.

## Example Usage

```terraform
data "scalr_hook" "example1" {
  id = "hook-xxxxxxxxxx"
}

data "scalr_hook" "example2" {
  name = "production"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) The identifier of the hook in the format `hook-<RANDOM STRING>`.
- `name` (String) The name of the hook.

### Read-Only

- `description` (String) Description of the hook.
- `interpreter` (String) The interpreter to execute the hook script, such as 'bash', 'python3', etc.
- `scriptfile_path` (String) Path to the script file in the repository.
- `vcs_provider_id` (String) ID of the VCS provider in the format `vcs-<RANDOM STRING>`.
- `vcs_repo` (List of Object) Settings for the repository where the hook script is stored. (see [below for nested schema](#nestedatt--vcs_repo))

<a id="nestedatt--vcs_repo"></a>
### Nested Schema for `vcs_repo`

Read-Only:

- `branch` (String)
- `identifier` (String)
