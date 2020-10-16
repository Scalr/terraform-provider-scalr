---
layout: "scalr"
page_title: "Provider: Scalr"
sidebar_current: "docs-scalr-index"
description: |-
  The Scalr provider is used to interact with the many resources supported by Scalr. The provider needs to be configured with the proper credentials before it can be used.
---

# Scalr Provider

## Example Usage

```hcl
# Configure the Scalr Provider
provider "scalr" {
  hostname = "${var.hostname}"
  token    = "${var.token}"
}

# Create a workspace
resource "scalr_workspace" "ws" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `hostname` - (Optional) The Scalr hostname to connect to.
  Defaults to `my.scalr.com`. Can be overridden by setting the
  `SCALR_HOSTNAME` environment variable.
* `token` - (Optional) The token used to authenticate with Scalr.
  Can be overridden by setting the `SCALR_TOKEN` environment variable.
