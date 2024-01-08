---
title: "scalr_current_account"
categorySlug: "scalr-terraform-provider"
slug: "provider_datasource_scalr_current_account"
parentDocSlug: "provider_datasources"
hidden: false
order: 3
---
## Data Source Overview

Retrieves the details of current account when using Scalr remote backend.

No arguments are required. The data source returns details of the current account based on the `SCALR_ACCOUNT_ID` environment variable that is automatically exported in the Scalr remote backend.

## Example Usage

```terraform
data "scalr_current_account" "account" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `id` (String) The identifier of the account.
- `name` (String) The name of the account.