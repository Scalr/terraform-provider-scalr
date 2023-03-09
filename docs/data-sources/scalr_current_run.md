
# Data Source `scalr_current_run` 

Allows you to get information about the current Terraform run when using a Scalr remote backend workspace, including VCS (Git) metadata.

## Example Usage

```hcl
data scalr_current_run example {
}
```

## Argument Reference

No arguments are required. The data source returns details of the current run based on the `SCALR_RUN_ID` shell variable that is automatically exported in the Scalr remoted backend.
If the shell variable is not present (e.g. during a local run) an empty run structure will be returned with the `id` attribute set to "-".

## Attribute Reference

* `id` - The ID of the run, in the format `run-<RANDOM STRING>`
* `environment_id` - The ID of the environment, in the format `env-<RANDOM STRING>`
* `workspace_name` - Workspace name.
* `vcs` - Contains details of the VCS configuration if the workspace is linked to a VCS repo.
* `is_destroy` - Boolean indicates if this is a "destroy" run.
* `is_dry` - Boolean indicates if this is a dry run, i.e. triggered by a Pull Request (PR). No apply phase if this is true.
* `message` - Message describing how the run was triggered
* `source` - The source of the run (VCS, API, Manual).

The `vcs` block contains:

* `repository_id` - ID of the VCS repo in the for `:org/:repo`.
* `branch` - The linked VCS repo branch.
* `commit` - Details of the last commit to the linked VCS repo.

The `vcs.commit` block contains:

* `message` - Message for the last commit.
* `sha` - SHA of the last commit.
* `author` - Details of the author of the last commit.

The `vcs.commit.author` block contains:

* `email` - email_address of the author in the VCS.
* `name` - Name of the author in the VCS.
* `username` - Username of the author in the VCS.
