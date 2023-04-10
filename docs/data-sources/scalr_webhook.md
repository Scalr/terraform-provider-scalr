
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
* `endpoint_id` - ID of the endpoint, in the format `ep-<RANDOM STRING>`.
* `environment_id` - ID of the environment, in the format `env-<RANDOM STRING>`.
* `events` - List of event IDs.
* `last_triggered_at` - Date/time when webhook was last triggered.
