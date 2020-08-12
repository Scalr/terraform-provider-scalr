---
layout: "scalr"
page_title: "Scalr: scalr_workspace"
sidebar_current: "docs-datasource-scalr-workspace-x"
description: |-
  Get information on a workspace.
---

# Data Source: scalr_workspace

Use this data source to get information about a workspace.

## Example Usage

```hcl
data "scalr_workspace" "test" {
  name           = "my-workspace-name"
  environment_id = "my-env"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the workspace.
* `environment_id` - (Required) ID of the environment.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspace's ID, which looks like
  `ws-<RANDOM STRING>`.
* `auto_apply` - Indicates whether to automatically apply changes when a
  Terraform plan is successful.
* `operations` - Indicates whether the workspace is using remote execution mode.
* `queue_all_runs` - Indicates whether all runs should be queued.
* `terraform_version` - The version of Terraform used for this workspace.
* `vcs_repo` - Settings for the workspace's VCS repository.
* `working_directory` - A relative path that Terraform will execute within.

The `vcs_repo` block contains:

* `identifier` - A reference to your VCS repository in the format `:org/:repo`
  where `:org` and `:repo` refer to the organization and repository in your VCS
  provider.
* `oauth_token_id` - OAuth token ID of the configured VCS connection.
