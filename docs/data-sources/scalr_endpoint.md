---
layout: "scalr"
page_title: "Scalr: scalr_endpoint"
sidebar_current: "docs-datasource-scalr-endpoint-x"
description: |-
  Get information on an endpoint.
---

# scalr_endpoint Data Source

This data source is used to retrieve details of an endpoint.

## Example Usage

```hcl
data "scalr_endpoint" "example" {
  id = "ep-xxxxxxxxxxx"
}
```

## Argument Reference

* `id` - (Required) Endpoint ID, in the format `ep-<RANDOM STRING>`.

## Attribute Reference

All arguments plus:

* `name` - Name of the endpoint.
* `secret_key` - Secret key to sign the webhook payload. 
* `url` - Endpoint URL. 
* `max_attempts` - Max delivery attempts of the payload. 
* `timeout` - Endpoint timeout (in seconds). 
* `environment_id` - ID of the environment, in the format `env-<RANDOM STRING>`