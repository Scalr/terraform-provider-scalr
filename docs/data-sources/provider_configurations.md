---
title: "scalr_provider_configurations"
categorySlug: "scalr-terraform-provider"
slug: "provider_datasource_scalr_provider_configurations"
parentDocSlug: "provider_datasources"
hidden: false
order: 13
---
## Data Source Overview

Retrieves a list of provider configuration ids by name or type.

## Example Usage

```terraform
data "scalr_provider_configurations" "aws" {
  name = "in:aws_dev,aws_demo,aws_prod"
}

data "scalr_provider_configurations" "google" {
  provider_name = "google"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `account_id` (String) The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.
- `name` (String) The query used in a Scalr provider configuration name filter.
- `provider_name` (String) The name of a Terraform provider.

### Read-Only

- `id` (String) The ID of this resource.
- `ids` (List of String) The list of provider configuration IDs, in the format [`pcfg-xxxxxxxxxxx`, `pcfg-yyyyyyyyy`].