
# Resource `scalr_policy_group`

Manage the state of policy groups in Scalr. Create, update and destroy.

## Example Usage

```hcl
resource "scalr_policy_group" "example" {
  name            = "instance_types"
  opa_version     = "0.29.4"
  account_id      = "acc-xxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
    path       = "policies/instance"
    branch     = "dev"
  }
}
```

## Argument Reference

* `name` - (Required) The name of a policy group.
* `account_id` - (Required) The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.
* `vcs_provider_id` - (Required) The identifier of a VCS provider, in the format `vcs-<RANDOM STRING>`.
* `vcs_repo` - (Required) Object. The VCS meta-data to create the policy from:

    * `identifier` - (Required) The reference to the VCS repository in the format `:org/:repo`, refers to the organization and repository in your VCS provider.
    * `branch` - (Optional) The branch of a repository the policy group is associated with. If omitted, the repository default branch will be used.
    * `path` - (Optional) The subdirectory of the VCS repository where OPA policies are stored. If omitted or submitted as an empty string, this defaults to the repository's root.

* `opa_version` - (Optional) The version of Open Policy Agent to run policies against. If omitted, the system default version is assigned.

## Attribute Reference

All arguments plus:

* `id` - An identifier of the policy group.
* `status` - A system status of the Policy group.
* `error_message` - A detailed error if Scalr failed to process the policy group.
* `policies` - A list of the OPA policies the group verifies each run.
* `environments` - A list of the environments the policy group is linked to.
* `workspaces` - A list of the workspaces this policy group verifies runs for.

The `policies` list contains definitions of OPA policies in the following form:

* `name` - A name of the policy.
* `enabled` - If set to `false`, the policy will not be verified during a run.
* `enforced_level` - An enforcement level of the policy.

## Import

To import policy groups use the policy group ID as the import ID. For example:

```shell
terraform import scalr_policy_group.example pgrp-svsu2dqfvtk5qfg
```
