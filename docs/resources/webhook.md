---
layout: "scalr"
page_title: "Scalr: scalr_webhook"
sidebar_current: "docs-resource-scalr-webhook"
description: |-
  Manages webhooks.
---

# scalr_webhook

Manage the state of webhooks in Scalr. Creates, updates and destroy.

## Example Usage

Basic usage:

```hcl
resource "scalr_webhook" "test" {
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
* `enabled` - (Optional) Whether webhook is enabled. 
* `endpoint_id` - (Required) ID of the endpoint, in the format `ep-<RANDOM STRING>`.
* `workspace_id` - (Optional) ID of the workspace, in the format `ws-<RANDOM STRING>`.
* `environment_id` - (Required if workspace ID is empty) ID of the environment, in the format `env-<RANDOM STRING>`.
* `events` - (Required) List of event IDs.

## Attributes

All arguments plus:

* `id` - The webhook ID, in the format `wh-<RANDOM STRING>`.
* `last_triggered_at` - Date/time when webhook was triggered last time.
