
# Resource `scalr_policy_group_linkage` 

Manage policy group to environment linking in Scalr. Create, update and destroy.

## Example Usage

```hcl
resource "scalr_policy_group_linkage" "example" {
  policy_group_id = "pgrp-xxxxxxxx"
  environment_id  = "env-xxxxxxxx"
}
```

## Argument Reference

* `policy_group_id` - (Required) ID of the policy group, in the format `pgrp-<RANDOM STRING>`.
* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - The ID of the policy group linkage.

## Import

To import policy group linkage use combined ID in the form `<policy_group_id>/<environment_id>` as the import ID. For example:

```shell
terraform import scalr_policy_group_linkage.example pgrp-tne44l0u69rmrm8/env-svrdqa8d7mhaimo
```
