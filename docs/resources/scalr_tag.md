---
layout: "scalr"
page_title: "Scalr: scalr_tag"
sidebar_current: "docs-resource-scalr-tag"
description: |-
  Manages tags.
---

# scalr_tag Resource

Manage the state of tags in Scalr.

## Example Usage

Basic usage:

```hcl
resource "scalr_tag" "example" {
  name       = "tag-name"
  account_id = "acc-<id>"
}
```

## Argument Reference

* `name` - (Required) Name of the tag.
* `account_id` - (Required) ID of the environment account, in the format `acc-<RANDOM STRING>`

## Attributes

All arguments plus:

* `id` - The identifier of the tag in the format `tag-<RANDOM STRING>`.