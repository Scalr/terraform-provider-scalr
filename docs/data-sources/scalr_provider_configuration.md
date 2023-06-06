
# Data Source `scalr_provider_configuration` 

Retrieves information about a single provider configuration.

## Example Usage

```hcl
data "scalr_provider_configuration" "aws_dev" {
  id = "pcfg-xxxxxxx"
}
```

```hcl
data "scalr_provider_configuration" "aws_dev" {
  name = "aws_dev"
}
```

```hcl
data "scalr_provider_configuration" "azure" {
  provider_name = "azurerm"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The provider configuration ID, in the format `pcfg-xxxxxxxxxxx`.
* `name` - (Optional) The name of a Scalr provider configuration.
* `provider_name` - (Optional) The name of a Terraform provider.
* `account_id` - (Optional) The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

Attributes are the same as arguments.
