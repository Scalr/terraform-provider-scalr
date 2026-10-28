---
name: provider-sdk-to-framework
description: >-
  MUST be used when migrating a Scalr provider resource or data source from
  Plugin SDK v2 (resource_scalr_*.go / data_source_scalr_*.go) to the Terraform
  Plugin Framework, or when moving one from go-scalr v1 to v2, either as a task
  of its own or as part of another task that touches the resource. The goal is
  zero user-visible change: same type name, same state, empty plan after upgrade.
---

# Legacy migration: SDK v2 → Plugin Framework, go-scalr v1 → v2

Migrations happen gradually, usually inside a task that touches the resource (new attribute,
bug fix). The developer decides whether to migrate; see
[AGENTS.md § Legacy resources](../../../AGENTS.md#legacy-resources).

A migration succeeds when a user upgrading the provider sees an empty plan. Everything below
serves that.

Inside another task: migrate first with no behaviour change, in its own commit, and get the
existing and upgrade tests passing; then make the task's change on top. If the port hits logic that does not
fit the Framework, or go-scalr v2 does not cover an endpoint correctly, stop and tell the
developer. Falling back to a narrow change on the legacy code and a separate migration ticket is
a valid outcome.

Past SDK → Framework migrations, for their migration mechanics (several still import v1 in the
resource or tests, so do not copy their imports): `tag_resource.go` +
`TestAccScalrTag_UpgradeFromSDK`, `workspace_*` (the largest, with state upgraders), `variable_*`
(write-only value, upgraders from SDK versions).

## Framework resource on go-scalr v1 → v2

A smaller job: the schema, state and Framework code stay; only the client calls change. The new
client code follows [provider-new-resource § CRUD rules](../provider-new-resource/SKILL.md#crud-rules).

1. Check the pinned v2 covers every operation, filter, include and field the resource uses
   (`go doc github.com/scalr/go-scalr/v2/scalr/ops/<name>`, `.../schemas`). v2 is generated
   from the spec, so expect gaps: generator bugs, unsupported spec constructs, a spec newer than
   the generator. Start with `make generate` in `../go-scalr/v2` and pin the merged commit
   ([AGENTS.md § Regenerating go-scalr v2](../../../AGENTS.md#regenerating-go-scalr-v2)). If
   that is not possible in this task, stop and tell the developer; the resource stays on v1.
2. Replace `r.Client` with `r.ClientV2`. Build requests with the tri-state `value` package and
   decide `Unset` vs `Null` per attribute so updates send the same fields v1 sent. v1 often sent
   zero values or omitted fields through `omitempty`; reproduce that behaviour, not the struct.
3. Not-found: `client.ErrNotFound` instead of `scalr.ErrResourceNotFound`.
4. Map responses with `types.StringPointerValue` etc. A v1 field that was `""` and is now `nil`
   in v2 changes state from `""` to `null`: that is a plan diff for existing users. Keep the old
   value or record it in the changelog `### Changed`.
5. Tests: switch imports and `CheckDestroy` to v2 (`createScalrClientV2()`). Existing steps must
   pass unchanged. No schema version change, so no `UpgradeFromSDK` test is needed; add a step
   with `plancheck.ExpectEmptyPlan()` after apply if mapping changed.
6. Remove the files from the depguard exception list in `.golangci.yml`; `golangci-lint run`
   must pass.
7. A file not named `<name>_resource.go` / `<name>_data_source.go` after the type name (e.g.
   `scalr_checkov_integration_resource.go`) is renamed with `git mv` in the same change.

The sections below are for SDK v2 → Framework. An SDK resource always lands on v2 directly.

## 1. Inventory the SDK resource

From `resource_scalr_<name>.go` (and its `_migrate.go`) write down, per attribute: type,
`Required`/`Optional`/`Computed`, `Default`, `ForceNew`, `Sensitive`, `Deprecated`,
`DiffSuppressFunc`, `ValidateFunc`, `ConflictsWith`/`ExactlyOneOf`, and whether it is a block
(`TypeList`/`TypeSet` of `&schema.Resource`). Also: `SchemaVersion`, `StateUpgraders`, the
importer, `CustomizeDiff`, timeouts.

Each item maps to a Framework construct. `DiffSuppressFunc` and `CustomizeDiff` become plan
modifiers or `ModifyPlan`; `Default` becomes a `Default`; `ForceNew` becomes `RequiresReplace()`.

## 2. Keep state compatible

- Type name: the Framework resource returns the same `scalr_<name>` in `Metadata`.
- Schema version: keep the SDK `SchemaVersion` as the Framework `Version`, or bump it and add
  `UpgradeState` entries from every SDK version (see `provider-schema-change` § State upgrades).
- Blocks: SDK nested blocks (`TypeList` + `MaxItems: 1`, `TypeSet` of resources) are stored as
  lists/sets of objects. Keep them as Framework `Blocks` (`schema.ListNestedBlock`,
  `SetNestedBlock`) so config syntax (`block { }`) and state shape stay the same. Converting to
  nested attributes changes the HCL users write: that is a breaking change.
- SDK stores unset strings as `""` and unset numbers as `0`; the Framework stores `null`. Decide
  per attribute and make Read produce what the plan will compute, or every upgraded user gets a
  diff. Where the SDK stored `""`, the Framework attribute can be `Optional`+`Computed` with a
  plan modifier, or the change is accepted and recorded in the changelog `### Changed`.
- Sets of primitives stay sets; lists stay lists.
- Importer ID format stays the same.

## 3. Swap the registration

- Remove from `provider_sdk.go` `ResourcesMap`/`DataSourcesMap`; add to `provider.go`. A type name
  registered in both providers makes the mux fail at startup.
- New file names drop the legacy prefix: `resource_scalr_<name>.go` → `<name>_resource.go`,
  `data_source_scalr_<name>.go` → `<name>_data_source.go`, `resource_scalr_<name>_migrate.go` →
  `<name>_migrate.go`, and the same for `_test.go`. `git mv` the test files before editing them
  so history follows the rename.
- A large resource, or one with SDK schema versions to keep, is a good candidate for the
  `_schemas.go` / `_models.go` / `_migrate.go` split (as `workspace_*` did); a small one stays in
  `<name>_resource.go`.
- Delete the SDK files and helpers only they used. Keep `_migrate.go` logic if it is still needed
  by an upgrader.
- The migrated code is new code: Framework + go-scalr v2 only
  ([AGENTS.md invariant 1](../../../AGENTS.md#architecture-invariants)), written to
  [provider-new-resource § Schema rules and CRUD rules](../provider-new-resource/SKILL.md#schema-rules).
  If v2 lacks an operation, add it to go-scalr v2 first. Check that v2 sends the same fields the
  SDK code did (step 2 of the v1 → v2 section applies here too).
- Remove from the depguard exception list every file this migration moved to v2.

## 4. Tests

- Port the existing acceptance tests to `ProtoV5ProviderFactories: protoV5ProviderFactories(t)`.
  Keep their configs; a migration does not change behaviour. Rewrite their imports to
  `terraform-plugin-testing` and go-scalr v2 (`createScalrClientV2()` in `CheckDestroy`).
- Add `TestAcc<Name>_UpgradeFromSDK`: step 1 applies the config with
  `ExternalProviders: {"scalr": {Source: "registry.scalr.io/scalr/scalr", VersionConstraint: "<=<last SDK release>"}}`,
  step 2 plans the same config with the current provider and asserts
  `plancheck.ExpectEmptyPlan()`. Use a config that sets every attribute, including blocks and
  optional values left unset. The last SDK release for the resource is the newest tag before this
  migration (`git tag --sort=-v:refname | head`).
- Unit-test pure state upgraders.
- Docs: `go generate`. The page changes only in formatting; a changed description, default or
  required-ness means the schema drifted from the SDK one. Fix the schema, not the doc.

## 5. Changelog

A clean migration has no user-visible change and needs no entry. Record any behaviour change it
introduces (`""` → `null`, stricter validation, new plan-time errors) under `### Changed`.

## Report

The inventory table, every intentional difference from the SDK behaviour, the schema version
before and after, and whether the `UpgradeFromSDK` test ran.
