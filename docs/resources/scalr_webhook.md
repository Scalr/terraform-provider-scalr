
# Resource `scalr_webhook`

Manage the state of webhooks in Scalr. Creates, updates and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_webhook" "example" {
  name           = "my-webhook-name"
  enabled        = true
  endpoint_id    = "ep-xxxxxxxxxx"
  events         = ["run:completed", "run:errored"]
  workspace_id   = "ws-xxxxxxxxxx"
  environment_id = "env-xxxxxxxxxx"
}
```

## Argument Reference

* `name` - (Required) Name of the webhook.
* `account_id` - (Optional) ID of the account, in the format `acc-<RANDOM STRING>`.
* `enabled` - (Optional) Set (true/false) to enable/disable the webhook. 
* `endpoint_id` - (Deprecated) ID of the endpoint, in the format `ep-<RANDOM STRING>`.
* `workspace_id` - (Deprecated) ID of the workspace, in the format `ws-<RANDOM STRING>`.
* `environment_id` - (Deprecated) ID of the environment, in the format `env-<RANDOM STRING>`.
* `events` - (Required) List of event IDs.
* `url` - (Optional) Endpoint URL. Required if `endpoint_id` is not set.
* `secret_key` - (Optional) Secret key to sign the webhook payload.
* `max_attempts` - (Optional) Max delivery attempts of the payload.
* `timeout` - (Optional) Endpoint timeout (in seconds).
* `environments` - (Optional) The list of environment identifiers that the webhook is shared to.
Use `["*"]` to share with all environments.
* `header` - (Optional, set of header objects) Additional headers to set in the webhook request.
  The `header` block item contains:
  * `name` - The name of the header.
  * `value` - The value of the header.

## Attributes

All arguments plus:

* `id` - The webhook ID, in the format `wh-<RANDOM STRING>`.
* `last_triggered_at` - Date/time when webhook was last triggered.
