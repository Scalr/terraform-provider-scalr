---
layout: "scalr"
page_title: "Scalr: scalr_endpoint"
sidebar_current: "docs-datasource-scalr-endpoint-x"
description: |-
  Get information on an endpoint.
---

# scalr_endpoint

This data source is used to retrieve details of an endpoint.

## Example Usage

```hcl
data "scalr_endpoint" "test" {
  id = "ep-xxxxxxxxxxx"
}
```

## Arguments

* `id` - (Required) Endpoint ID.

## Attributes

All arguments plus:

* `id` - The endpoint's ID, in the format `ep-<RANDOM STRING>`.
* `name` - Name of the endpoint.
* `secret_key` - Secret key to sign payload. 
* `url` - Endpoint URL. 
* `max_attempts` - Max delivery attempts of the payload. 
* `timeout` - Endpoint timeout (in sec). 
* `environment_id` - ID of the environment, in the format `env-<RANDOM STRING>`