---
layout: "scalr"
page_title: "Scalr: scalr_endpoint"
sidebar_current: "docs-datasource-scalr-endpoint-x"
description: |-
  Get information on an endpoint.
---

# Data Source: scalr_endpoint

Use this data source to get information about an endpoint.

## Example Usage

```hcl
data "scalr_endpoint" "test" {
  id = "my-endpoint-ID"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) Endpoint ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The endpoint's ID, which looks like `ep-<RANDOM STRING>`.
* `name` - Name of the endpoint.
* `secret_key` - Secret key to sign payload. 
* `url` - Endpoint URL. 
* `max_attempts` - Max delivery attempts. 
* `timeout` - Endpoint timeout (in sec). 
* `environment_id` - ID of the environment.
