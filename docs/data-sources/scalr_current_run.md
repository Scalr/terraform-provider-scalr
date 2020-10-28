---
layout: "scalr"
page_title: "Scalr: scalr_current_run"
sidebar_current: "docs-datasource-scalr-current-run-x"
description: |-
  Get information on the current run.
---

# scalr_current_run Data Source

This data source allows you to get information about the current Terraform run when using a Scalr remote backend workspace, including VCS (Git) metadata.

## Example Usage

```hcl
data scalr_current_run example {
  }
```

## Argument Reference

No arguments required. This data source returns details of the current run.

## Attribute Reference

All arguments plus:

* `id` - The ID of the run, in the format `run-<RANDOM STRING>`
* `environment_id` - The ID of the environment, in the format `env-<RANDOM STRING>`
* `workspace_name` - Workspace name.
* `vcs` - Contains details of the VCS configuration if the workspace is linked to a VCS repo.
* `is_destroy` - Boolean indicates if this is a "destroy" run.
* `is_dry` - Boolean indicates if this is a dry run, i.e. triggered by a Pull Request (PR). No apply phase if this is true.
* `message` - Message describing how the run was triggered
* `source` - The source of the run (VCS, api, Manual).

The `vcs` block contains:

* `repository_id` - ID of the VCS repo in the for `:org/:repo`.
* `branch` - The linked VCS repo branch.
* `commit` - Details of the last commit to the linked VCS repo.

The `vcs.commit` block contains:

* `message` - Message for the last commit.
* `sha` - SHA of the last commit.
* `author` - Details of the commit author.

The `vcs.commit.author` block contains:

* `email` - email_address of author in the VCS.
* `name` - Name of author in the VCS.
* `username` - Username of author in the VCS.