<!-- TOC -->
* [How to contribute to the Scalr provider](#how-to-contribute-to-the-scalr-provider)
  * [Updating the changelog](#updating-the-changelog)
    * [Added](#added)
    * [Changed](#changed)
    * [Fixed](#fixed)
    * [Deprecated](#deprecated)
    * [Removed](#removed)
  * [Local setup](#local-setup)
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

New resources and datasources are built on the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
and the [Scalr Go Client v2](https://github.com/Scalr/go-scalr/tree/master/v2) [^1].
Legacy code is migrated gradually: the Plugin SDK v2 resources (`resource_scalr_*.go`, `data_source_scalr_*.go`) to the Framework,
and Framework resources still using the v1 client to the v2 client. The v1 client will be removed entirely.
When a task touches a legacy resource (a new attribute, a bug fix), consider migrating it as part of the task. This is preferred, not required:
a small or urgent fix is better kept narrow, and a resource with logic that does not port cleanly to the Framework,
or that uses endpoints the generated v2 client does not support correctly yet, is better migrated in a separate task.
Gaps in the v2 client are fixed in [go-scalr](https://github.com/Scalr/go-scalr) (spec or generator), not worked around with v1.
When migrating, do it first without behaviour changes, in its own commit, then make the actual change on top.

> [!IMPORTANT]
> New resources and datasources, including their tests, must not import the Plugin SDK (`github.com/hashicorp/terraform-plugin-sdk/v2`)
  or the v1 client (`github.com/scalr/go-scalr`). Tests use `github.com/hashicorp/terraform-plugin-testing` and `createScalrClientV2()`.
  Many existing Framework resources and tests still use the v1 client; do not copy them.
  The `depguard` linter in [.golangci.yml](./.golangci.yml) enforces this. Its list of legacy exceptions may only shrink.

The easiest way to start is to copy the closest recent resource and adapt it, together with its test file. These use the Framework and the v2 client throughout:
- `internal/provider/datadog_integration_resource.go` and `datadog_integration_data_source.go`: an account-scoped object with a secret attribute
- `internal/provider/docker_integration_resource.go`, `wiz_integration_resource.go`: similar integrations
- `internal/provider/workspace_remote_state_consumer_resource.go`: a link between two objects with a composite import ID

A large resource can be split into `<name>_resource.go` (CRUD), `<name>_schemas.go` (current and prior schema versions),
`<name>_models.go` (model structs and API mapping) and `<name>_migrate.go` (state upgraders). It is optional: a simple resource reads
better in one file. For an example of the split and of state upgraders, see `internal/provider/workspace_{resource,schemas,models,migrate}.go`
(its test file still uses the v1 client).

The main steps to follow when adding a resource are:
- make sure the API is available in the Scalr Go Client v2; if not, regenerate it: with `fatmouse`, `go-scalr` and this repository checked out
  side by side, run `make generate` in `go-scalr/v2` (it reads `../../fatmouse/taco/openapi/openapi-public.yml`), and merge the result in go-scalr [^1].
  While iterating, `go mod edit -replace github.com/scalr/go-scalr/v2=../go-scalr/v2` lets you build against the local client; drop it before pushing
- pin `go-scalr` dependency to proper commit: `go get github.com/scalr/go-scalr/v2@<commit-sha> && go mod tidy`
- add `internal/provider/<name>_resource.go`, `internal/provider/<name>_data_source.go`, implement schemas and methods.
  `<name>` is the type name without `scalr_`. The legacy `resource_scalr_<name>.go` / `data_source_scalr_<name>.go` names
  are not used for new code, and a migrated resource is renamed to the new scheme

> [!IMPORTANT]
> Always fill in the `MarkdownDescription` field for the resource/datasource schema and for every attribute in it
  with clean and useful information. This will be collected and compiled into the documentation website.
- register new resources and datasources in `Resources()` and `DataSources()` in [internal/provider/provider.go](./internal/provider/provider.go)
- add corresponding `*_test.go` files for each new module with acceptance tests
- add the example files for the documentation, see the `examples` folder for the reference
- if the resource needs a more complex doc page, this can be done by adding a new template in the `templates` folder
- run `go generate` command from repository root to compile the documentation
- [update the changelog](#updating-the-changelog)

Contributors using AI coding agents: see [AGENTS.md](./AGENTS.md) and the skills in `.agents/skills/`.

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

Example:

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

Scalr is SaaS only, so the changelog does not state a minimum Scalr server version. Older releases have a `Required` section; new ones do not.

When in doubt refer to [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## Local setup

- Go version from [.go-version](./.go-version), and the `terraform` binary (used by `go generate` and the `terraform fmt` hook).
- [golangci-lint v2](https://golangci-lint.run/) at the version CI uses: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.1`.
- [pre-commit](https://pre-commit.com/) (`brew install pre-commit` or `pipx install pre-commit`), then once per clone:
  ```shell
  pre-commit install
  ```
  The hooks format Go code (gofumpt and gci import groups, see [.golangci.yml](./.golangci.yml)), lint it,
  check `go.mod`/`go.sum` are tidy, run `terraform fmt` on examples and fix trailing whitespace and final newlines.
  CI runs the same checks. `make fmt` formats all Go code.

## Guidelines

Refer to https://developer.hashicorp.com/terraform/plugin/best-practices for the base rules.

### Naming conventions

Resource and datasource names in provider schema should follow the pattern `scalr_<name>`, where `<name>` usually represents the resource name as it defined in the API specification (examples: `scalr_workspace`, `scalr_tag`).

Use combined term for `<name>` when the resource represents an objects relation or a separate property, as in: `scalr_environment_hook`, `scalr_workspace_var_set`, `scalr_provider_configuration_default`, `scalr_workspace_run_schedule`.

A resource that links two other resources (like a many-to-many table in a database) is named after the two resources it connects:
`scalr_<resource1>_<resource2>`, the object it is attached to first: `scalr_environment_hook` attaches a hook to an environment.
Do not add `_link`, `_linkage` or similar suffixes; `scalr_policy_group_linkage` predates this rule and keeps its name, as existing type names never change.

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
- always cleanup `go.sum` after modifying project dependencies (`go get`, bumping `go-scalr`, removing a local `replace`),
  so it does not keep outdated entries for a dependency next to the newer ones:
  ```shell
  go mod tidy
  go mod tidy -diff   # prints nothing when go.mod and go.sum are clean; CI runs the same check
  ```

## Common pitfalls

### Working with to-many relationships

When a resource schema contains an attribute that presents a to-many relationship, there is a risk to accidently break it if not being carefull. Such relationship, when implementing the create/update structs in go-scalr client, should **NOT** be marked with `omitempty` tag. Otherwise, JSONAPI marshaller will not include an empty slice in the payload so there will be no way to set the relationship to an empty value.

But having client to send the relationship payload even when empty leads to another problem - when implementing the Update operation for the resource, we should always populate to-many attributes with existing values even when they are not changed, or it will accidently clear the value by sending an empty value. Always double check the to-many attributes work correctly on resource update.

### Circular dependencies

Considering a scenario where there are two resources and each resource has a relationship to another (one or both of these relationships are optional, otherwise it's more of an API design issue), there is no way to create both in a single terraform run, as it forms a circular dependency where each resource requires another to be created before it.

It can be solved by making one of these attributes `Computed` and introducing a new resource that manages the state of this relation,
named after the two resources it connects (see [Naming conventions](#naming-conventions)). Refer to `scalr_provider_configuration`, `scalr_provider_configuration_default` and `scalr_environment.default_provider_configurations` for an example.

[^1]: The v2 client is generated from the Scalr OpenAPI spec. The v1 client (repository root of `go-scalr`) is hand-written and frozen; it is only extended for resources that still use it.
[^2]: Some advices are relevant only for SDKv2-based resources; these are being migrated to the Terraform Plugin Framework
