
# Data Source `scalr_endpoint` 

Retrieves the details of a webhook endpoint.

## Example Usage

```hcl
data "scalr_endpoint" "example" {
  id = "ep-xxxxxxxxxxx"
}
```

```hcl
data "scalr_endpoint" "example" {
  name = "endpoint_name"
}
```

```hcl
data "scalr_endpoint" "example" {
  name = "endpoint_name"
}
```

## Argument Reference

* `id` - (Optional) The endpoint ID, in the format `env-<RANDOM STRING>`.
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