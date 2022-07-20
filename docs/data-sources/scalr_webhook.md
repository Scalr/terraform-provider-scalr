---
layout: "scalr"
page_title: "Scalr: scalr_webhook"
sidebar_current: "docs-datasource-scalr-webhook-x"
description: |-
  Get information on a webhook.
---

# scalr_webhook Data Source

This data source is used to retrieve details of a webhook.

## Example Usage

```hcl
data "scalr_webhook" "example" {
  id = "wh-xxxxxxxxxxx"
}
```

## Argument Reference

* `id` - (Optional) The webhook ID, in the format `wh-<RANDOM STRING>`.
* `name` - (Optional) Name of the webhook.
* `environment_id` - (Optional) ID of the webhook environment, in the format `env-<RANDOM STRING>`

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_webhook`.

## Attribute Reference

All arguments plus:

* `enabled` - Boolean indicates if the webhook is enabled. 
* `endpoint_id` - ID of the endpoint, in the format `ep-<RANDOM STRING>`.
* `workspace_id` - ID of the workspace if applicable, in the format `ws-<RANDOM STRING>`.
* `events` - List of event IDs.
* `last_triggered_at` - Date/time when webhook was last triggered.
