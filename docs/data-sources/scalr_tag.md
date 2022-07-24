---
layout: "scalr"
page_title: "Scalr: scalr_tag"
sidebar_current: "docs-datasource-scalr-tag-x"
description: |-
  Get information on a tag.
---

# scalr_tag Data Source

This data source is used to retrieve details of a tag.

## Example Usage

```hcl
data "scalr_tag" "example" {
  name = "tag-name"
  account_id = "acc-<id>"
}
```

## Arguments

* `name` - (Required) Name of the tag.
* `account_id` - (Required) ID of the environment account, in the format `acc-<RANDOM STRING>`

## Attributes

All arguments plus:

* `id` - The identifier of the tag in the format `tag-<RANDOM STRING>`.