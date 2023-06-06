
# Data Source `scalr_endpoint` 

Retrieves the details of a webhook endpoint.

> **WARNING:** This datasource is deprecated and will be removed in the next major version.

## Example Usage

```hcl
data "scalr_endpoint" "example" {
  id         = "ep-xxxxxxxxxxx"
  account_id = "acc-xxxxxxx"
}
```

```hcl
data "scalr_endpoint" "example" {
  name       = "endpoint_name"
  account_id = "acc-xxxxxxx"
}
```

## Argument Reference

* `id` - (Optional) The endpoint ID, in the format `ep-<RANDOM STRING>`.
* `name` - (Optional) Name of the endpoint.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_endpoint`.

## Attribute Reference

All arguments plus:

* `secret_key` - Secret key to sign the webhook payload. 
* `url` - Endpoint URL. 
* `max_attempts` - Max delivery attempts of the payload.
* `environment_id` - ID of the environment, in the format `env-<RANDOM STRING>`.
* `timeout` - Endpoint timeout (in seconds). 