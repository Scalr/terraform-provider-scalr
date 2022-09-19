
# Data Source `scalr_module_version`

Retrieves the module version data by module source and semantic version.

## Example Usage

```hcl
data "scalr_module_version" "example" {
  source = "env-xxxxxx/resource-name/scalr"
  version = "1.0.0"
}
```

## Argument Reference

The following arguments are supported:

* `source` - (Required) The module source.
* `version` - (Optional) The semantic version based on module version was created.

## Attribute Reference

All arguments plus:

* `id` - The identifier of а module version. Example: `modver-xxxx`
