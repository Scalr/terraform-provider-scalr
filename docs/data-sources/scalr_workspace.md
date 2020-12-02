---
layout: "scalr"
page_title: "Scalr: scalr_workspace"
sidebar_current: "docs-datasource-scalr-workspace-x"
description: |-
  Get information on a workspace.
---

# scalr_workspace Data Source

This data source is used to retrieve details of a single workspace by name.

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
* `vcs_provider_id` - (Optional) ID of vcs provider, in the format `vcs-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - The workspace's ID, in the format `ws-<RANDOM STRING>`.
* `auto_apply` - Boolean indicates if `terrafrom apply` will be automatically run when `terraform plan` ends without error.
* `operations` - Boolean indicates if the workspace is being used for remote execution.
* `terraform_version` - The version of Terraform used for this workspace.
* `working_directory` - A relative path that Terraform will execute within.
* `vcs_repo` - If workspace is linked to VCS repository this block shows the details, otherwise `{}`
* `created_by` - Details of the user that created the workspace.

The `vcs_repo` block contains:

* `identifier` - * The reference to the VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.
* `path` - Path within the repo, if any.

The `created_by` block contains:

* `username` - Username of creator.
* `email` - Email address of creator.
* `full_name` - Full name of creator.