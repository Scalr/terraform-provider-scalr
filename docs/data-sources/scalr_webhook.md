
# Data Source `scalr_webhook`

Retrieves the details of a webhook.

## Example Usage

```hcl
data "scalr_webhook" "example" {
  id         = "wh-xxxxxxxxxxx"
  account_id = "acc-xxxxxxx"
}
```

```hcl
data "scalr_webhook" "example" {
  name       = "webhook_name"
  account_id = "acc-xxxxxxx"
}
```

## Argument Reference

* `id` - (Optional) The webhook ID, in the format `wh-<RANDOM STRING>`.
* `name` - (Optional) Name of the webhook.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_webhook`.

## Attribute Reference

All arguments plus:

* `enabled` - Boolean indicates if the webhook is enabled. 
* `endpoint_id` - (Deprecated) ID of the endpoint, in the format `ep-<RANDOM STRING>`.
* `environment_id` - (Deprecated) ID of the environment, in the format `env-<RANDOM STRING>`.
* `workspace_id` - (Deprecated) ID of the workspace, in the format `ws-<RANDOM STRING>`.
* `events` - List of event IDs.
* `last_triggered_at` - Date/time when webhook was last triggered.
* `url` - Endpoint URL. 
* `secret_key` - Secret key to sign the webhook payload.
* `max_attempts` - Max delivery attempts of the payload.
* `timeout` - Endpoint timeout (in seconds).
* `environments` - The list of environment identifiers that the webhook is shared to,
or `["*"]` if shared with all environments.
* `header` - (Set of header objects) Additional headers to set in the webhook request.
  The `header` block item contains:
  * `name` - The name of the header.
  * `value` - The value of the header.
