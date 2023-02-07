
# Resource `scalr_provider_configuration_default`

Manage defaults or provider configurations for environments in Scalr. Create and destroy.

## Basic Usage

```hcl

resource "scalr_provider_configuration_default" "example" {
  environment_id = "env-xxxxxxxx"
  provider_configuration_id = "pcfg-xxxxxxxx"
}
    
    ```

## Argument Reference

* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.
* `provider_configuration_id` - (Required) ID of the provider configuration, in the format `pcfg-<RANDOM STRING>`. 

Note:
Provider configuration should be in list of environment identifiers that the provider configuration is shared to.

## Attribute Reference

All arguments plus:

* `id` - The ID of the provider configuration defaults.
  
## Import

To import provider configuration defaults use combined ID in the form `<environment_id>/<provider_configuration_id>` as the import ID. For example:

```shell

terraform import scalr_provider_configuration_defaults.example env-svrdqa8d7mhaimo/pcfg-xxxxxxxx

```
