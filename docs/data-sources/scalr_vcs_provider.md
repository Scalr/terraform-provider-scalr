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

* `vcs_type` - (Optional) Type of the VCS provider. For example, `github`.
* `name` - (Optional) Name of the VCS provider.
* `environment_id` - (Optional) ID of the environment the VCS provider has to be linked to, in the format `env-<RANDOM STRING>`.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `id` - Identifier of the VCS provider, in the format `vcs-<RANDOM STRING>`.
* `url` - The URL to the VCS provider installation.
* `environments` - List of the identifiers of environments the VCS provider is linked to.
