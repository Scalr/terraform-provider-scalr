---
layout: "scalr"
page_title: "Scalr: scalr_policy_group"
sidebar_current: "docs-resource-scalr-policy-group"
description: |-
  Manages policy groups.
---

# scalr_policy_group Resource

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

* `name` - (Required) Name of the policy group.
* `account_id` - (Required) ID of the account, in the format `acc-<RANDOM STRING>`.
* `vcs_provider_id` - (Required) ID of a VCS provider, in the format `vcs-<RANDOM STRING>`.
* `vcs_repo` - (Required) Settings for the policy group's VCS repository.

    The `vcs_repo` block supports:
    * `identifier` - (Required) The reference to the VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.
    * `branch` - (Optional) Branch of a repository the policy group is associated with. If omitted, the repository default branch will be used.
    * `path` - (Optional) The sub-directory of the VCS repository where OPA policies are stored. If omitted or submitted as an empty string, this defaults to the repository's root.

* `opa_version` - (Optional) The version of Open Policy Agent to use for the policy evaluation. If omitted, the system default version is assigned.

## Attribute Reference

All arguments plus:

* `id` - The ID of the policy group.
* `status` - Policy group current status.
* `error_message` - The error description when the group's status is `errored`.
* `policies` - List of OPA policies this group contains.
* `environments` - List of environments this policy group is linked to.
* `workspaces` - List of workspaces affected by this policy group.

The `policies` list contains definitions of OPA policies in the following form:

* `name` - The name of the policy.
* `enabled` - If set to `false`, the policy will not be evaluated during a run.
* `enforced_level` - The policy's enforcement level.

## Import

To import policy groups use policy group ID as the import ID. For example:

```shell
terraform import scalr_policy_group.example pgrp-svsu2dqfvtk5qfg
```
