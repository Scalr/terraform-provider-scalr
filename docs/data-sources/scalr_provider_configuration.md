
# Data Source `scalr_provider_configuration` 

Retrieves the id of a single provider configuration by name or type.

## Example Usage

```javascript
data "scalr_provider_configuration" "aws_dev" {
  name = "aws_dev"
}

data "scalr_provider_configuration" "azure" {
  provider_name = "azurerm"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of a Scalr provider configuration.
* `provider_name` - (Optional) The name of a Terraform provider.
* `account_id` - (Optional) The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - The provider configuration ID, in the format `pcfg-xxxxxxxxxxx`.