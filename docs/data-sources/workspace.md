---
title: "scalr_workspace"
categorySlug: "scalr-terraform-provider"
slug: "provider_datasource_scalr_workspace"
parentDocSlug: "provider_datasources"
hidden: false
order: 22
---
## Data Source: scalr_workspace

Retrieves the details of a single workspace.

## Example Usage

```terraform
data "scalr_workspace" "example1" {
  id             = "ws-xxxxxxxxxx"
  environment_id = "env-xxxxxxxxxx"
}

data "scalr_workspace" "example2" {
  name           = "my-workspace-name"
  environment_id = "env-xxxxxxxxxx"
}
```

<!-- Manually filling the schema here because of https://github.com/hashicorp/terraform-plugin-docs/issues/28 -->
## Schema

### Required

- `environment_id` (String) ID of the environment, in the format `env-<RANDOM STRING>`.

### Optional

- `id` (String) ID of the workspace.
- `name` (String) Name of the workspace.

### Read-Only

- `agent_pool_id` (String) The identifier of an agent pool in the format `apool-<RANDOM STRING>`.
- `auto_apply` (Boolean) Boolean indicates if `terraform apply` will be automatically run when `terraform plan` ends without error.
- `auto_queue_runs` (String) Indicates if runs have to be queued automatically when a new configuration version is uploaded.

  Supported values are `skip_first`, `always`, `never`:

  * `skip_first` - after the very first configuration version is uploaded into the workspace the run will not be triggered. But the following configurations will do. This is the default behavior.
  * `always` - runs will be triggered automatically on every upload of the configuration version.
  * `never` - configuration versions are uploaded into the workspace, but runs will not be triggered.
- `created_by` (List of Object) Details of the user that created the workspace. (see [below for nested schema](#nestedatt--created_by))
- `deletion_protection_enabled` (Boolean) Boolean, indicates if the workspace has the protection from an accidental state lost. If enabled and the workspace has resource, the deletion will not be allowed.
- `execution_mode` (String) Execution mode of the workspace.
- `force_latest_run` (Boolean) Boolean indicates if latest new run will be automatically raised in priority.
- `has_resources` (Boolean) The presence of active terraform resources in the current state version.
- `hooks` (List of Object) List of custom hooks in a workspace. (see [below for nested schema](#nestedatt--hooks))
- `module_version_id` (String) The identifier of a module version in the format `modver-<RANDOM STRING>`.
- `operations` (Boolean) Boolean indicates if the workspace is being used for remote execution.
- `tag_ids` (List of String) List of tag IDs associated with the workspace.
- `terraform_version` (String) The version of Terraform used for this workspace.
- `iac_platform` (String) The IaC platform used for this workspace.
- `vcs_provider_id` (String) The identifier of a VCS provider in the format `vcs-<RANDOM STRING>`.
- `vcs_repo` (List of Object) If a workspace is linked to a VCS repository this block shows the details, otherwise `{}` (see [below for nested schema](#nestedatt--vcs_repo))
- `working_directory` (String) A relative path that Terraform will execute within.

<a id="nestedatt--created_by"></a>
### Nested Schema for `created_by`

Read-Only:

- `email` (String) Email address of creator.
- `full_name` (String) Full name of creator.
- `username` (String) Username of creator.


<a id="nestedatt--hooks"></a>
### Nested Schema for `hooks`

Read-Only:

- `post_apply` (String) Script or action configured to call after apply phase.
- `post_plan` (String) Script or action configured to call after plan phase.
- `pre_apply` (String) Script or action configured to call before apply phase.
- `pre_init` (String) Script or action configured to call before init phase.
- `pre_plan` (String) Script or action configured to call before plan phase.


<a id="nestedatt--vcs_repo"></a>
### Nested Schema for `vcs_repo`

Read-Only:

- `dry_runs_enabled` (Boolean) Boolean indicates the VCS-driven dry runs should run when the pull request to the configuration versions branch is created.
- `identifier` (String) The reference to the VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.
- `ingress_submodules` (Boolean) Designates whether to clone git submodules of the VCS repository.
- `path` (String) Path within the repo, if any.
