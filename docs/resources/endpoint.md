---
title: "scalr_endpoint"
categorySlug: "scalr-terraform-provider"
slug: "provider_resource_scalr_endpoint"
parentDocSlug: "provider_resources"
hidden: false
order: 5
---
## Resource Overview

Manage the state of endpoints in Scalr. Create, update and destroy.

> 🚧 This resource is deprecated and will be removed in the next major version.

## Example Usage

```terraform
resource "scalr_endpoint" "example" {
  name           = "my-endpoint-name"
  secret_key     = "my-secret-key"
  timeout        = 15
  max_attempts   = 3
  url            = "https://my-endpoint.url"
  environment_id = "env-xxxxxxxxxx"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (String) ID of the environment, in the format `env-<RANDOM STRING>`.
- `name` (String) Name of the endpoint.
- `url` (String) Endpoint URL.

### Optional

- `max_attempts` (Number) Max delivery attempts.
- `secret_key` (String, Sensitive) Secret key to sign payload.
- `timeout` (Number) Endpoint timeout (in sec).

### Read-Only

- `id` (String) The ID of this resource.

## Useful snippets

The secret key can be generated using the `random_string` resource.

```terraform
resource "random_string" "r" {
  length = 16
}

resource "scalr_endpoint" "example" {
  # ...
  secret_key = random_string.r.result
  # ...
}
```

## Import

Import is supported using the following syntax:

```shell
terraform import scalr_endpoint.example ep-xxxxxxxxxx
```