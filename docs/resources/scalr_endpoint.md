---
layout: "scalr"
page_title: "Scalr: scalr_endpoint"
sidebar_current: "docs-resource-scalr-endpoint"
description: |-
  Manages endpoints.
---

# scalr_endpoint Resource

Manage the state of endpoints in Scalr. Create, update and destroy

## Example Usage

Basic usage:

```hcl
resource "scalr_endpoint" "example" {
  name           = "my-endpoint-name"
  secret_key     = "my-secret-key"
  timeout        = 15
  max_attempts   = 3
  url            = "https://my-endpoint.url"
  environment_id = "env-xxxxxxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) Name of the endpoint.
* `secret_key` - (Required) Secret key to sign payload. 
* `environment_id` - (Required) ID of the environment, in the format `env-<RANDOM STRING>`.
* `url` - (Required) Endpoint URL. 
* `max_attempts` - (Optional) Max delivery attempts. 
* `timeout` - (Optional) Endpoint timeout (in sec). 

## Attribute Reference

All arguments plus:

* `id` - The endpoint's ID, in the format `ep-<RANDOM STRING>`.

## Useful snippets

Secret key can be generated using the `random_string` resource.

```hcl
resource "random_string" "r" {
  length = 16
}

resource "scalr_endpoint" "example" {
  # ...
  secret_key = random_string.r.result
  # ...
}
```
