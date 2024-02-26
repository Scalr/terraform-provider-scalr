<!-- TOC -->
* [How to contribute to the Scalr provider](#how-to-contribute-to-the-scalr-provider)
  * [Updating the changelog](#updating-the-changelog)
    * [Added](#added)
    * [Changed](#changed)
    * [Removed](#removed)
    * [Required](#required)
  * [Guidelines](#guidelines)
    * [Naming conventions](#naming-conventions)
    * [Rules to follow](#rules-to-follow-3)
  * [Common pitfalls](#common-pitfalls)
    * [Working with to-many relationships](#working-with-to-many-relationships)
    * [Circular dependencies](#circular-dependencies)
<!-- TOC -->

# How to contribute to the Scalr provider

The most common task when developing the plugin is adding support for a new Scalr resource.

Usually this means implementing a terraform resource with CRUD operations and a corresponding datasource (the datasource can be skipped if there are no evident use cases for it).

The main steps to follow when adding a resource are:
- implement resource structs and CRUD methods in [Scalr Go Client](https://github.com/Scalr/go-scalr/) [^1]
- pin `go-scalr` dependency to proper commit: `go get github.com/Scalr/go-scalr@<commit-sha>` [^1]
- add `scalr/resource_scalr_<name>.go`, `scalr/datasource_scalr_<name>.go`, implement schemas and methods
  
> [!IMPORTANT]
> Always fill in the `Description` field for the resource/datasource schema and for every attribute in it
  with clean and useful information. This will be collected and compiled into the documentation website.
- add new resources to [provider schema](./scalr/provider.go)
- add corresponding `*_test.go` files for each new module with acceptance tests
- add the example files for the documentation, see the `examples` folder for the reference
- if the resource needs a more complex doc page, this can be done by adding a new template in the `templates` folder
- run `go generate` command from repository root to compile the documentation
- [update the changelog](#updating-the-changelog)

## Updating the changelog

Each PR should update the [CHANGELOG.md](./CHANGELOG.md) file to include all changes affecting provider behaviour or compatibility. The version and date heading should be **Unreleased** and updated after being merged and released on the main branch.

The format of the changelog for each release looks like this:

```
## [Unreleased]

### Added
...
### Changed
...
### Fixed
...
### Deprecated
...
### Removed
...
### Required
...
```

> [!NOTE]
> Changing single attribute is just as essential as modifying the resource or datasource.
  Therefore any new, removed or deprecated attributes go to the corresponding 'Added', 'Removed' or 'Deprecated' section.
  For example, it's easy to confuse adding new attribute with only 'changing' the resource, however it should belong to the 'Added' section, as it introduces a new feature.

### <a name="added"></a>Added: for new features

In this section we need to mention any resources, attributes or other features that have been added. This is generally a list of new things the user can do after this release. Example:

```
- `scalr_workspace`: new attribute `environment_id` (Scalr environment ID, replaces `organization`) ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))
```

### <a name="changed"></a>Changed: for changes in existing functionality

Existing resources, attributes or behaviour which changes in this release. Attribute changes are likely to break compatibility with older provider configs, however we should never break the state. If any attribute changes in this release it needs to include a state migration and a test for it.

Sometimes there will be changes in go-scalr which may affect the minimum required version of Scalr. In such case we can give a quick overview of changes in go-scalr without going too deep into its internals. Example:

```
- `scalr_workspace`: attribute `id` is now in the `ws-<RANDOM STRING>` format ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))
```

### <a name="fixed"></a>Fixed: for any bug fixes

Any changes that fix the issues in existing functionality. Example:

```
- `scalr_webhook`: fix handling resource destroy when resource no longer exists
```

### <a name="deprecated"></a>Deprecated: for soon-to-be removed features

Mention any functionality that is considered now deprecated. Include any useful information on deprecation period, migration tips, etc. Example:

```
- `scalr_endpoint` is deprecated and will be removed in the next major version. The endpoint information is included in the `scalr_webhook` resource.
```

### <a name="removed"></a>Removed: for now removed features

This section should detail any resources, attributes or features removed from the provider in this version. This generally means the user will need to make changes to their configs after updating. Example:

```
- `scalr_workspace`: drop attribute `organization` in favour of `environment_id` ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))
```

### <a name="required"></a>Required: server version dependency

If the provider relies on changes introduced in a more recent version of Scalr we should specify its version here. If the provider can be used with an older version of Scalr there is no need to update this version from the previous release.

```
- scalr server >= `8.63.0`
```

When in doubt refer to [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## Guidelines

Refer to https://developer.hashicorp.com/terraform/plugin/best-practices for the base rules.

### Naming conventions

Resource and datasource names in provider schema should follow the pattern `scalr_<name>`, where `<name>` usually represents the resource name as it defined in the API specification (examples: `scalr_workspace`, `scalr_tag`).

Use combined term for `<name>` when the resource represents an objects relation or a separate property, as in: `scalr_policy_group_linkage`, `scalr_provider_configuration_default`, `scalr_workspace_run_schedule`.

The datasource name should be in a singular form for datasource that returns a single object (usually a simple lookup by id or name), and in plural for the objects listing (usually should accept some filtering or search options). Examples: `scalr_environment` and `scalr_environments`.

Resource attributes are usually named as they appear in the API, replacing kebab-case with a snake_case form (`is-system` -> `is_system`).

API object relationships in provider resource schema usually follow the rule:
- to-one relationship is presented as `<name>_id attribute`: `environment->relationships->account` becomes `environment->account_id` in provider
- to-many relationship go as it is (in plural): `webhook->relationships->environments` becomes `webhook->environments`

When in doubt, it is always a good advice to look through some good examples of existing terraform plugins, such as [terraform-provider-aws](https://github.com/hashicorp/terraform-provider-aws), [terraform-provider-auth0](https://github.com/auth0/terraform-provider-auth0), [terraform-provider-tfe](https://github.com/hashicorp/terraform-provider-tfe)

### Rules to follow [^2]

- avoid using deprecated features
- use context-aware function versions where it is intended (resource CRUD, etc),
as well as Diagnostics-enabled functions
- prefer declaring validation logic within the schema that can trigger earlier in Terraform operations,
rather than using create or update logic which only triggers during apply
- take advantage of a `diag.Diagnostics` type. Return multiple errors and warnings to Terraform
  where it fits, associate those errors or warnings with specific fields
- use `schema.Resource.Description` and `schema.Schema.Description` to document the resources and fields
- when error is not handled on purpose, ignore it explicitly so that it does not produce
  inspection warning, and informs others this was intentional, e.g.:
    ```go
    _ = d.Set("message", run.Message)
    ```
- error check aggregate types: when setting value for a non-primitive type (`TypeList`, `TypeSet`, `TypeMap`) check the result of `d.Set()` for an error. Aggregate types are converted to key/value pairs when set into state, and if it's not checked for error, Terraform will think it's operation was successful despite having broken state. This rule is not currently followed in the codebase, but it should be
- be consistent declaring and initiating variables:
  - `var a int` when declaring the variable
  - `a := 1` walrus notation when declaring and initializing the variable
  - when a slice must be initialized with an empty slice instead of zero value,
    prefer allocating it with `make` function instead of empty slice literal (`make([]int, 0)` over `[]int{}`)

> [!NOTE]
> When choosing the initial value for slice, take into account that zero-value slice marshals into `null`,
  while an empty slice will produce `[]`.
- always cleanup `go.sum` after modifying project dependencies:
  ```shell
  go mod tidy
  ```

## Common pitfalls

### Working with to-many relationships

When a resource schema contains an attribute that presents a to-many relationship, there is a risk to accidently break it if not being carefull. Such relationship, when implementing the create/update structs in go-scalr client, should **NOT** be marked with `omitempty` tag. Otherwise, JSONAPI marshaller will not include an empty slice in the payload so there will be no way to set the relationship to an empty value.

But having client to send the relationship payload even when empty leads to another problem - when implementing the Update operation for the resource, we should always populate to-many attributes with existing values even when they are not changed, or it will accidently clear the value by sending an empty value. Always double check the to-many attributes work correctly on resource update.

### Circular dependencies

Considering a scenario where there are two resources and each resource has a relationship to another (one or both of these relationships are optional, otherwise it's more of an API design issue), there is no way to create both in a single terraform run, as it forms a circular dependency where each resource requires another to be created before it.

It can be solved by making one of these attributes `Computed` and introducing a new resource that manages the state of this relation. Refer to `scalr_provider_configuration`, `scalr_provider_configuration_default` and `scalr_environment.default_provider_configurations` for an example.

[^1]: Approach will change in the future when the autogenerated API client will be introduced
[^2]: Some advices are relevant only for SDKv2-based plugin; the migration to the Terraform Plugin Framework is already scheduled and this doc will be continiously updated and supplemented accordingly
