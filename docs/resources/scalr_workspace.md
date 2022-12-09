
# Resource `scalr_workspace` 

Manage the state of workspaces in Scalr. Create, update and destroy

## Example Usage

Basic usage:

### VCS-driven
```hcl
data "scalr_vcs_provider" test {
  name = "vcs-name"
  account_id = "acc-xxxx" # in case if user has access to more than one account
}

data "scalr_environment" test {
  name = "env-name"
  account_id = "acc-xxxx" # in case if a user has access to more than one account
}

resource "scalr_workspace" "vcs-driven" {
  name            = "my-workspace-name"
  environment_id  = data.scalr_environment.test.id
  vcs_provider_id = data.scalr_vcs_provider.test.id

  working_directory = "example/path"

  vcs_repo {
      identifier          = "org/repo"
      branch              = "dev"
      trigger_prefixes    = ["stage", "prod"]
  }

  provider_configuration {
    id    = "pcfg-1"
    alias = "us_east1"
  }
  provider_configuration {
    id    = "pcfg-2"
    alias = "us_east2"
  }
}
```

### Module-driven

```hcl
data "scalr_environment" test {
  name = "env-name"
  # account_id = "acc-xxxx" # Optional, in case if user has access to more than one account
}

locals {
  modules = {
    "${data.scalr_environment.test.id}": "module-name/provider",         # environment-level module will be selected
    "${data.scalr_environment.test.account_id}": "module-name/provider", # account-level module will be selected
  }
}

data "scalr_module_version" "module-driven" {
  for_each = local.modules
  source = "${each.key}/${each.value}"
  # version = "1.0.0" # Optional, if omitted, the latest module version is selected
}

resource "scalr_workspace" "example" {
  for_each = data.scalr_module_version.example
  environment_id = data.scalr_environment.test.id

  name = replace(each.value.source, "/", "-")
  module_version_id = each.value.id
}
```

### CLI-driven

```hcl
data "scalr_environment" test {
  name = "env-name"
  # account_id = "acc-xxxx" # Optional, in case if user has access to more than one account
}

resource "scalr_workspace" "cli-driven" {
  name            = "my-workspace-name"
  environment_id  = data.scalr_environment.test.id

  working_directory = "example/path"
}
```

## Argument Reference

* `name` - (Required) Name of the workspace.
* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.
* `auto_apply` - (Optional) Set (true/false) to configure if `terraform apply` should automatically run when `terraform plan` ends without error. Default `false`.
* `force_latest_run` - (Optional) Set (true/false) to configure if latest new run will be automatically raised in priority. Default `false`.
* `operations` - Deprecated. Use `execution-mode` instead.
* `auto_queue_runs` - Indicates if runs have to be queued automatically when a new configuration version is uploaded.

  Supported values are `skip_first`, `always`, `never`:

  * `skip_first` - after the very first configuration version is uploaded into the workspace the run will not be triggered. But the following configurations will do. This is the default behavior.
  * `always` - runs will be triggered automatically on every upload of the configuration version.
  * `never` - configuration versions are uploaded into the workspace, but runs will not be triggered.

* `execution-mode` - (Optional) Which execution mode to use. Valid values are `remote` and `local`. When set to `local`, the workspace will be used for state storage only. Defaults to `remote` (not set, backend default is used).
* `terraform_version` - (Optional) The version of Terraform to use for this workspace. Defaults to the latest available version.
* `working_directory` - (Optional) A relative path that Terraform will be run in. Defaults to the root of the repository `""`.
* `var_files` - (Optional) A list of paths to the `.tfvars` file(s) to be used as part of the workspace configuration.
* `run_operation_timeout` - (Optional) The number of minutes run operation can be executed before termination. Defaults to `0` (not set, backend default is used).
* `module_version_id` - (Optional) The identifier of a module version in the format `modver-<RANDOM STRING>`. This attribute conflicts with `vcs_provider_id` and `vcs_repo` attributes.
* `agent_pool_id` - (Optional) The identifier of an agent pool in the format `apool-<RANDOM STRING>`.
* `tag_ids` - (Optional) List of tag IDs associated with the workspace.
* `vcs_provider_id` - (Optional) ID of vcs provider - required if vcs-repo present and vice versa, in the format `vcs-<RANDOM STRING>`
* `vcs_repo` - (Optional) Settings for the workspace's VCS repository.

    The `vcs_repo` block supports: 
    * `identifier` - (Required) A reference to your VCS repository in the format `:org/:repo`, it refers to the organization and repository in your VCS provider.
    * `branch` - (Optional) The repository branch where Terraform will be run from. Default `master`.
    * `path` - (Optional) `Deprecated`: The repository subdirectory that Terraform will execute from. If omitted or submitted as an empty string, this defaults to the repository's root.
    * `trigger_prefixes` - (Optional) List of paths (relative to `path`), whose changes will trigger a run for the workspace using this binding when the CV is created. If omitted or submitted as an empty list, any change in `path` will trigger a new run.
    * `dry_runs_enabled` - (Optional) Set (true/false) to configure the VCS driven dry runs should run when pull request to configuration versions branch created. Default `true`
    * `ingress_submodules` - (Optional) Designates whether to clone git submodules of the VCS repository.

* `hooks` - (Optional) Settings for the workspaces custom hooks.

   The `hooks` block supports: 
  * `pre_init` - (Optional) Action that will be called before the init phase
  * `pre_plan` - (Optional) Action that will be called before the plan phase
  * `post_plan` - (Optional) Action that will be called after plan phase
  * `pre_apply` - (Optional) Action that will be called before apply phase
  * `post_apply` - (Optional) Action that will be called after apply phase

* `provider_configuration` - (Optional) Provider configurations used in workspace runs.

   The `provider_configuration` block supports:
  * `id` - (Required) The identifier of provider configuration
  * `alias` - (Optional) The alias of provider configuration

## Attribute Reference

All arguments plus:

* `id` - The workspace ID, in the format `ws-<RANDOM STRING>`.
* `created_by` - Details of the user that created the workspace.
* `has_resources` - The presence of active terraform resources in the current state version.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.

## Import

To import workspaces use workspace ID as the import ID. For example:

```shell
terraform import scalr_workspace.example ws-t47s1aa6s4boubg
```
