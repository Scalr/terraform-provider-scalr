
# Data Source `scalr_workspace`

Retrieves the details of a single workspace by name.

## Example Usage

```hcl
data "scalr_workspace" "example" {
  name           = "my-workspace-name"
  environment_id = "env-xxxxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the workspace.
* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - The workspace ID, in the format `ws-<RANDOM STRING>`.
* `auto_apply` - Boolean indicates if `terraform apply` will be automatically run when `terraform plan` ends without error.
* `force_latest_run` - Boolean indicates if latest new run will be automatically raised in priority.
* `operations` - Boolean indicates if the workspace is being used for remote execution.
* `execution-mode` - Execution mode of the workspace.
* `terraform_version` - The version of Terraform used for this workspace.
* `working_directory` - A relative path that Terraform will execute within.
* `run_operation_timeout` - The number of minutes run operation can be executed before termination.
* `module_version_id` - The identifier of a module version in the format `modver-<RANDOM STRING>`.
* `tag_ids` - List of tag IDs associated with the workspace.
* `vcs_provider_id` - The identifier of a VCS provider in the format `vcs-<RANDOM STRING>`.
* `vcs_repo` - If a workspace is linked to a VCS repository this block shows the details, otherwise `{}`
* `created_by` - Details of the user that created the workspace.
* `has_resources` - The presence of active terraform resources in the current state version.
* `auto_queue_runs` - Indicates if runs have to be queued automatically when a new configuration version is uploaded. Supported values are `default` (only on VCS changes), `enabled`, `disabled`.
* `hooks` - List of custom hooks in a workspace.

  The `hooks` block supports:

  * `pre_init` - Script or action configured to call before init phase
  * `pre_plan` - Script or action configured to call before plan phase
  * `post_plan` - Script or action configured to call after plan phase
  * `pre_apply` - Script or action configured to call before apply phase
  * `post_apply` - Script or action configured to call after apply phase

The `vcs_repo` block contains:

* `identifier` - * The reference to the VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.
* `path` - Path within the repo, if any.
* `dry_runs_enabled` - Boolean indicates the VCS-driven dry runs should run when the pull request to the configuration versions branch is created.
* `ingress_submodules` - Designates whether to clone git submodules of the VCS repository.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.
