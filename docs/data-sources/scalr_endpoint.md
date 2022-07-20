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

* `id` - (Optional) The endpoint ID, in the format `env-<RANDOM STRING>`.
* `name` - (Optional) Name of the endpoint.
* `environment_id` - (Optional) ID of the endpoint environment, in the format `env-<RANDOM STRING>`

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_endpoint`.

## Attribute Reference

All arguments plus:

* `secret_key` - Secret key to sign the webhook payload. 
* `url` - Endpoint URL. 
* `max_attempts` - Max delivery attempts of the payload. 
* `timeout` - Endpoint timeout (in seconds). 