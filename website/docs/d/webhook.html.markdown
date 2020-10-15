---
layout: "scalr"
page_title: "Scalr: scalr_webhook"
sidebar_current: "docs-datasource-scalr-webhook-x"
description: |-
  Get information on a webhook.
---

# scalr_webhook

This data source is used to retrieve details of a webhook.

## Example Usage

```hcl
data "scalr_webhook" "test" {
  id = "my-webhook-ID"
}
```

## Arguments

The following arguments are supported:

* `id` - (Required) Webhook ID.

## Attributes

All arguments plus:

* `id` - The webhook ID, in the format `wh-<RANDOM STRING>`.
* `name` - Name of the webhook.
* `enabled` - Whether webhook is enabled. 
* `endpoint_id` - ID of the endpoint.
* `workspace_id` - ID of the workspace if applicable.
* `environment_id` - ID of the environment.
* `events` - List of event IDs.
* `last_triggered_at` - Date/time when webhook was triggered last time.
