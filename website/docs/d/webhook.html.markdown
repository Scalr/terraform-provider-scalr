---
layout: "scalr"
page_title: "Scalr: scalr_webhook"
sidebar_current: "docs-datasource-scalr-webhook-x"
description: |-
  Get information on a webhook.
---

# Data Source: scalr_webhook

Use this data source to get information about a webhook.

## Example Usage

```hcl
data "scalr_webhook" "test" {
  id = "my-webhook-ID"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) Webhook ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The webhook's ID, which looks like `wh-<RANDOM STRING>`.
* `name` - Name of the webhook.
* `enabled` - Whether webhook is enabled. 
* `endpoint_id` - ID of the endpoint.
* `workspace_id` - ID of the workspace.
* `environment_id` - ID of the environment.
* `events` - List of event IDs.
* `last_triggered_at` - Datetime when webhook was triggered last time.
