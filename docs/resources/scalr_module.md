---
layout: "scalr"
page_title: "Scalr: scalr_module"
sidebar_current: "docs-resource-scalr-module"
description: |-
  Manages module.
---

# scalr_module Resource

Manage the state of module in Scalr. Create and destroy

## Example Usage

Basic usage:

```hcl
resource "scalr_module" "example" {
  account_id      = "acc-xxxxxxxxx"
  environment_id  = "env-xxxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxxx"
  vcs_repo {
      identifier          = "org/repo"
      branch              = "dev"
      path                = "example/terraform-<provider>-<name>"
      tag_prefix          = "aws/"
  }
}

```

## Argument Reference
* `vcs_provider_id` - (Required) ID of vcs provider, in the format `vcs-<RANDOM STRING>`
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.
* `environment_id` - (Optional) ID of the environment, in the format `env-<RANDOM STRING>`.
* `vcs_repo` - (Optional) Settings for the module's VCS repository.

    The `vcs_repo` block supports:
 
    * `identifier` - (Required) A reference to your VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.
    * `path` - (Optional) The path to the root module folder. Is expected to have the format '../terraform-<provider_name>-<module_name>'
    * `tag_prefix` - (Optional) Registry ignores tags which do not match specified prefix, e.g. `aws/`.
    

## Attribute Reference

All arguments plus:

* `id` - The module's ID, in the format `mod--<RANDOM STRING>`.
* `module_provider` - Module provider retrieved from the vcs_repo.identifier.
* `name` - Module name retrieved from the vcs_repo.identifier.
* `source` - The remote module source that should be used in terraform templates

## Import

To import module use module ID as the import ID. For example:
```shell
terraform import scalr_module.example mod-tk4315k3lofu4i0
```
