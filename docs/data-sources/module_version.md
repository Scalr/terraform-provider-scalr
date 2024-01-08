---
title: "scalr_module_version"
categorySlug: "scalr-terraform-provider"
slug: "provider_datasource_scalr_module_version"
parentDocSlug: "provider_datasources"
hidden: false
order: 10
---
## Data Source Overview

Retrieves the module version data by module source and semantic version.

## Example Usage

```terraform
data "scalr_module_version" "example" {
  source  = "env-xxxxxxxxxx/resource-name/scalr"
  version = "1.0.0"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `source` (String) The module source.

### Optional

- `version` (String) The semantic version. If omitted, the latest module version is selected

### Read-Only

- `id` (String) The identifier of а module version. Example: `modver-xxxx`