---
layout: "scalr"
page_title: "Scalr: scalr_workspace"
sidebar_current: "docs-resource-scalr-workspace"
description: |-
  Manages workspaces.
---

# scalr_workspace Resource

Manage the state of workspaces in Scalr. Create, update and destroy

## Example Usage

Basic usage:

```hcl
resource "scalr_workspace" "example" {
  name            = "my-workspace-name"
  environment_id  = "env-xxxxxxxxx"
  vcs_provider_id = "my_vcs_provider"
  vcs_repo {
      identifier          = "org/repo"
      branch              = "dev"
  }
}
```

## Argument Reference

* `name` - (Required) Name of the workspace.
* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.
* `auto_apply` - (Optional) Set (true/false) to configure if `terraform apply` should automatically run when `terraform plan` ends without error. Default `false`.
* `operations` - (Optional) Set (true/false) to configure workspace remote execution. When `false` workspace is only used to store state. Default `true`.
  Defaults to `true`.
* `terraform_version` - (Optional) The version of Terraform to use for this workspace. Defaults to the latest available version.
* `working_directory` - (Optional) A relative path that Terraform will be run in. Defaults to the root of the repository.
* `vcs_provider_id` - (Optional) ID of vcs provider - required if vcs-repo present and vice versa, in the format `vcs-<RANDOM STRING>`
* `vcs_repo` - (Optional) Settings for the workspace's VCS repository.

The `vcs_repo` block supports:

* `identifier` - (Required) A reference to your VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.
* `branch` - (Optional) The repository branch where Terraform will be run from. Default `master`.

## Attribute Reference

All arguments plus:

* `id` - The workspace's ID, in the format `ws-<RANDOM STRING>`.
* `created_by` - Details of the user that created the workspace.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.

## Import

To import workspaces use workspace ID as the import ID. For example:
```shell
terraform import scalr_workspace.example ws-t47s1aa6s4boubg
```
