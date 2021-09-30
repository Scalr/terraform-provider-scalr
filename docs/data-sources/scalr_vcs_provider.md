---
layout: "scalr"
page_title: "Scalr: scalr_vcs_provider"
sidebar_current: "docs-datasource-scalr-vcs-provider-x"
description: |-
  Get information on a vcs provider.
---

# scalr_vcs_provider Data Source

This data source is used to retrieve details of a vcs_provider.

## Example Usage

```hcl
data "scalr_vcs_provider" "example" {
  id = "vcs-xxxxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) Identifier of the vcs provider, in the format `vcs-<RANDOM STRING>`. 
* `vcs_type` - (Optional) Type of the vcs provider. For example, `github`.
* `name` - (Optional) Name of the vcs provider.
* `environment` - (Optional) ID of the environment the vcs provider has to be linked to, in the format `env-<RANDOM STRING>`.
* `account` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `environments` - List of the identifiers of environments the vsc provider is linked to.
