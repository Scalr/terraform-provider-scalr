
# Resource `scalr_module`

Manages the state of a module in the Private Modules Registry. Create and destroy operations are available only.

## Example Usage

Basic usage:

```hcl
resource "scalr_module" "example" {
  account_id      = "acc-xxxxxxxxx"
  environment_id  = "env-xxxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxxx"
  vcs_repo {
      identifier          = "org/repo"
      path                = "example/terraform-<provider>-<name>"
      tag_prefix          = "aws/"
  }
}

```

## Argument Reference
* `vcs_provider_id` - (Required) The identifier of a VCS provider in the format `vcs-<RANDOM STRING>`
* `account_id` - (Optional) The identifier of the account in the format `acc-<RANDOM STRING>`. If it is not specified the module will be registered globally and available across the whole installation.
* `environment_id` - (Optional) The identifier of an environment in the format `env-<RANDOM STRING>`. If it is not specified the module will be registered at the account level and available across all environments within the account specified in `account_id` attribute.
* `vcs_repo` - (Required) Source configuration of a VCS repository

    The `vcs_repo` block supports:
 
    * `identifier` - (Required) The identifier of a VCS repository in the format `:org/:repo` (`:org/:project/:name` is used for Azure DevOps). It refers to an organization and a repository name in a VCS provider.
    * `path` - (Optional) The path to the root module folder. It Is expected to have the format '<path>/terraform-<provider_name>-<module_name>', where `<path>` stands for any folder within the repository inclusively a repository root.
    * `tag_prefix` - (Optional) Registry ignores tags which do not match specified prefix, e.g. `aws/`.
    

## Attribute Reference

All arguments plus:

* `id` - The identifier of a module in the format `mod--<RANDOM STRING>`.
* `module_provider` - Module provider name, e.g `aws`, `azurerm`, `google`, etc.
* `name` - Name of the module, e.g. `rds`, `compute`, `kubernetes-engine`

* `source` - The source of a remote module in the private registry, e.g `env-xxxx/aws/vpc`

## Import

To import a module use the module ID as the import ID. For example:

```shell
terraform import scalr_module.example mod-tk4315k3lofu4i0
```
