---
layout: "scalr"
page_title: "Scalr: scalr_webhook"
sidebar_current: "docs-resource-scalr-webhook"
description: |-
  Manages webhooks.
---

# scalr_webhook

Provides a webhook resource.

## Example Usage

Basic usage:

```hcl
resource "scalr_webhook" "test" {
  name           = "my-webhook-name"
  enabled        = true
  endpoint_id    = "my-endpoint-id"
  events         = ["run:completed", "run:errored"]
  workspace_id   = "my-workspace-ID"
  environment_id = "my-environment-ID"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the webhook.
* `enabled` - (Optional) Whether webhook is enabled. 
* `endpoint_id` - (Required) ID of the endpoint.
* `workspace_id` - (Optional) ID of the workspace.
* `environment_id` - (Required if workspace ID is empty) ID of the environment.
* `events` - (Required) List of event IDs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The webhook's ID, which looks like `wh-<RANDOM STRING>`.
* `last_triggered_at` - Datetime when webhook was triggered last time.
