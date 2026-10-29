---
name: provider-schema-change
description: >-
  MUST be used when adding, changing, deprecating or removing an attribute or
  block on an existing Scalr provider resource or data source, or when a change
  could alter stored state (type, mode, default, ID format, nesting). Covers
  resource/data-source parity, SDK vs Framework, state upgraders, plan
  stability, docs and changelog.
---

# Attribute and schema changes

## 1. Locate every surface

An attribute usually exists in several places. Find all before editing:

```bash
grep -rn '"<attr_name>"' internal/provider/ examples/ templates/
```

- Resource and its data source (`scalr_x` and `data.scalr_x`). New readable attributes go to both
  unless there is a reason not to; say which when you skip one.
- Plural data sources (`scalr_environments`, `scalr_workspaces`, ...) that embed the same fields.
- Split files, when the resource is split: `<name>_schemas.go`, `<name>_models.go`
  (`...FromAPI`), `<name>_resource.go` (Create/Update request building),
  `<name>_migrate.go`. A first schema version bump on a single-file resource is a reasonable
  point to suggest the split; ask first.
- SDK v2 resources (`resource_scalr_*.go`): schema map, `Create`/`Update` option structs, `Read`
  `d.Set` calls.

If any surface is legacy (an SDK v2 resource or data source, or a Framework one still on
go-scalr v1), ask the developer now whether to migrate it as part of this task, with a
recommendation ([AGENTS.md § Legacy resources](../../../AGENTS.md#legacy-resources)). Check that
the pinned go-scalr v2 covers the new attribute either way. On a yes, run
`provider-sdk-to-framework` first, then continue here on the migrated version.

## 2. Classify the change

| Change | Compatible? | Needs |
|---|---|---|
| New `Optional` or `Computed` attribute | Yes | Description, request + read mapping, test step |
| New `Optional` + `Computed` with server default | Yes | `UseStateForUnknown()` if the value is stable, test that omitting it yields no diff |
| New `Required` attribute | No: breaks configs | Ask first. Usually make it `Optional` |
| Change type, nesting, or `Optional`↔`Computed`↔`Required` | No | Ask first; schema version bump + state upgrader |
| Add `RequiresReplace` to an existing attribute | Behaviour change | Ask first; changelog `### Changed` |
| Change how an absent value is stored (`""` → `null`) | State shape | Schema version bump unless Read rewrites it with no plan diff; changelog `### Changed` |
| Rename | No | Add new, deprecate old, keep both working, remove in a major version |
| Deprecate | Yes | `DeprecationMessage` naming the replacement; changelog `### Deprecated` |
| Remove | No | Only in a major release after deprecation; changelog `### Removed` |

Server-side requirement: if the attribute needs fatmouse work that is not deployed yet, note the
fatmouse branch or PR; the provider is released only after it is deployed.

## 3. Implement

Framework code follows [provider-new-resource § Schema rules and CRUD
rules](../provider-new-resource/SKILL.md#schema-rules). For an attribute change in particular:

- Description first: what it does, allowed values, default, and that it is read-only where
  relevant.
- Update: when the attribute is removed from config, decide explicitly between leaving the server
  value (`value.Unset`) and clearing it (`value.Null`). Most optional attributes must clear.
- Read is also what import uses: an attribute missing from Read breaks `ImportStateVerify`.
- Data source: add to schema, model and Read mapping.
- SDK v2: `d.Set` errors for list/set/map attributes must be checked.

## 4. State upgrades

Needed when stored state from an older provider version cannot be read by the new schema, or
reads with a different meaning.

Framework (`workspace_migrate.go`, `variable_migrate.go`, `role_migrate.go` are references):

1. Bump `Version` in the schema.
2. Keep a frozen copy of each prior schema (`<name>ResourceSchemaV<N>` in `<name>_schemas.go`)
   and a local prior model struct inside the upgrader.
3. Add `UpgradeState(ctx) map[int64]resource.StateUpgrader` entries (`PriorSchema` +
   `StateUpgrader`) from every prior version straight to the current one (`V0toV4`, `V1toV4`,
   ...). The Framework does not chain them, so on a bump every existing upgrader must also emit
   the new current shape; rename them to the new target version.
4. Prefer pure transformation. Calling the API from an upgrader (as the workspace one does) is
   acceptable only when the old state lacks data that cannot be derived.

SDK v2: `SchemaVersion` + `StateUpgraders` with a `Type` from the previous schema; unit test with
`assertCorrectState` (see `resource_scalr_vcs_provider_migrate_test.go`).

Tests: a unit test for pure upgraders, and an acceptance step that applies with the last released
provider (`ExternalProviders` with `VersionConstraint`) then plans with the current one expecting
`plancheck.ExpectEmptyPlan()` (see `TestAccScalrWorkspaceResource_UpgradeFromSDK`).

## 5. Tests

Extend the resource's existing test functions with steps rather than new files:

- Set the attribute, update it, remove it from config (assert server value and state:
  `TestCheckNoResourceAttr` or the default).
- Omitting it on an existing resource produces no plan diff.
- Invalid values fail at plan time (`PlanOnly: true`, `ExpectError`).
- Import step still passes `ImportStateVerify`.
- Data source exposes it.

## 6. Docs and changelog

- Update `examples/` when the attribute is something users commonly set.
- `go generate`; the resource and data source pages change.
- Changelog per [provider-changelog](../provider-changelog/SKILL.md): one line for `scalr_x`, one
  for `data.scalr_x`.

## Report

List every surface touched and skipped (with reason), the compatibility class from the table, the
state version if bumped, and which acceptance tests ran.
