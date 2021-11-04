---
layout: "scalr"
page_title: "Scalr: scalr_policy_group"
sidebar_current: "docs-datasource-scalr-policy-group-x"
description: |-
  Get information on a policy group.
---

# scalr_policy_group Data Source

This data source is used to retrieve details of a policy group by name and account_id.

## Example Usage

```hcl
data "scalr_policy_group" "example" {
  name           = "instance_types"
  account_id     = "acc-xxxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) Name of the policy group.
* `account_id` - (Required) ID of the account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - The ID of the policy group.
* `status` - Policy group current status.
* `error_message` - The error description when the group's status is `errored`.
* `opa_version` -  The version of Open Policy Agent to use for the policy evaluation.
* `vcs_provider_id` - The identifier of a VCS provider in the format `vcs-<RANDOM STRING>`.
* `vcs_repo` - Contains details of the VCS configuration of the policy group.
* `policies` - List of OPA policies this group contains.
* `environments` - List of environments this policy group is linked to.
* `workspaces` - List of workspaces affected by this policy group.

The `vcs_repo` block contains:

* `identifier` - The reference to the VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.
* `branch` - Branch of a repository the policy group is associated with.
* `path` - The sub-directory of the VCS repository where OPA policies are stored.

The `policies` list contains definitions of OPA policies in the following form:

* `name` - The name of the policy.
* `enabled` - If set to `false`, the policy will not be evaluated during a run.
* `enforced_level` - The policy's enforcement level.
