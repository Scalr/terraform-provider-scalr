---
layout: "scalr"
page_title: "Scalr: scalr_policy_group"
sidebar_current: "docs-datasource-scalr-policy-group-x"
description: |-
  Get information on a policy group.
---

# scalr_policy_group Data Source

Retrieves the details of a policy group by the name and account_id.

## Example Usage

```hcl
data "scalr_policy_group" "example" {
  name           = "instance_types"
  account_id     = "acc-xxxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) The name of a policy group.
* `account_id` - (Required) The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - An identifier of the policy group.
* `status` - A system status of the policy group.
* `error_message` - An error details if Scalr failed to process the policy group.
* `opa_version` -  A version of Open Policy Agent the policy group verify run on. 
* `vcs_provider_id` - An identifier of a VCS provider the policy group is create from. In the format `vcs-<RANDOM STRING>`.
* `vcs_repo` - Contains VCS-related meta-data the policy group is created from.
* `policies` - A list of the OPA policies the policy group verifies each run.
* `environments` - A list of the environments the policy group is linked to.
* `workspaces` - A list of the workspaces this policy group verify runs for.

The `vcs_repo` object contains:

* `identifier` - A reference to the VCS repository in the format `:org/:repo`, it stands for the organization and repository.
* `branch` - A branch of a repository the policy group is associated with.
* `path` - A subdirectory of a VCS repository where OPA policies are stored.

A `policies` list contains definitions of OPA policies in the following form:

* `name` - A name of a policy.
* `enabled` - If set to `false`, the policy will not be verified on a run.
* `enforced_level` - An enforcement level of a policy.
