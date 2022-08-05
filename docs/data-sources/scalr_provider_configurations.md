
# Data Source `scalr_provider_configurations` 

Retrieves a list of provider configuration ids by name or type.

## Example Usage

```javascript
data "scalr_provider_configurations" "aws" {
  name = "in:aws_dev,aws_demo,aws_prod"
}

data "scalr_provider_configurations" "google" {
  provider_name = "google"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The query used in a Scalr provider configuration name filter.
* `provider_name` - (Optional) The name of a Terraform provider.

## Attribute Reference

All arguments plus:

* `ids` - The list of provider configuration IDs, in the format [`pcfg-xxxxxxxxxxx`, `pcfg-yyyyyyyyy`].