---
name: provider-new-resource
description: >-
  MUST be used when adding a new resource or data source to the Scalr provider
  (scalr_<name>), e.g. "add scalr_x resource", "support the new X integration",
  "expose X as a data source". Covers go-scalr v2 availability, the
  starting point, schema, CRUD, import, registration, tests, examples, docs and changelog.
  For attributes on an existing resource use provider-schema-change.
---

# New resource or data source

## Before writing code

1. Read the API: the fatmouse OpenAPI spec (`../fatmouse/taco/openapi/openapi-public.yml`) or the
   go-scalr v2 schema `../go-scalr/v2/scalr/schemas/<name>.gen.go` and ops
   `../go-scalr/v2/scalr/ops/<name>/`. Note which attributes are required, read-only, write-only
   (secrets the API never returns), and which relationships are to-one or to-many.
2. Check the client in the pinned version: `go doc github.com/scalr/go-scalr/v2/scalr/ops/<name>`.
   Missing or stale (the usual case for a new API) → first step is `make generate` in
   `../go-scalr/v2`; follow
   [AGENTS.md § Regenerating go-scalr v2](../../../AGENTS.md#regenerating-go-scalr-v2).
3. If the API lives only on a fatmouse branch, this branch must have the same name (Jira key) so
   CI acceptance tests get a Scalr with that API.
4. Decide the shape with the user when unclear:
   - Name: `scalr_<api_name>`; a relation/property gets a combined name
     (`scalr_workspace_remote_state_consumer`). A link resource is named after the two
     resources it connects, the object it attaches to first: `scalr_environment_hook`,
     `scalr_workspace_var_set`. No `_link` / `_linkage` suffix (`scalr_policy_group_linkage` is
     legacy). Data source: singular for lookup, plural for lists.
   - Lookup keys for the data source (`id` and/or `name`, `ExactlyOneOf`/`AtLeastOneOf`).
   - Import ID format (plain ID, or `<parent_id>/<child_id>` for link resources).
   - Circular relationships: make one side computed and add a separate link resource
     ([CONTRIBUTING.md § Circular dependencies](../../../CONTRIBUTING.md#circular-dependencies)).

## Allowed imports

Framework + go-scalr v2 only, tests included
([AGENTS.md invariant 1](../../../AGENTS.md#architecture-invariants), enforced by depguard). If v2
lacks an operation the resource or its `CheckDestroy` needs, add it to v2 (step 2 above). Never
add a new file to the depguard exception list to get past the linter.

## Start from a neighbour

Many Framework resources and tests still use go-scalr v1; do not copy those. Copy the closest of
these, which are Framework + v2 throughout, tests included, and adapt it:
`datadog_integration_resource.go` / `_data_source.go` (account-scoped object with a secret),
`docker_integration_resource.go`, `wiz_integration_resource.go`, or
`workspace_remote_state_consumer_resource.go` (link resource with a composite ID). Take the
test file from the same neighbour. Files: `internal/provider/<name>_resource.go`,
`<name>_data_source.go`, and their `_test.go`, with `<name>` the type name without `scalr_`. A
large resource may be split into `_schemas.go` / `_models.go` from the start; a small one stays
in one file ([AGENTS.md § Code Conventions](../../../AGENTS.md#code-conventions)).

## Schema rules

These and the CRUD rules apply to any Framework resource code: new resources, migrations, and
changes to existing Framework resources.

- Compile-time interface assertions at the top of the file
  (`var _ resource.ResourceWithImportState = &xResource{}`).
- Reuse `internal/framework` before writing a helper, plan modifier or validator; put a new
  shared one there, not in `internal/provider`. Existing:
  - plan modifiers (`planmodifiers/stringmodifier`): `UseStateOrNullForUnknown`,
    `UseStateForUnknownUnlessChanged(paths...)`;
  - validators: `stringvalidation.StringIsNotWhiteSpace`, `StringIsValidURL`,
    `StringIsNamespaceName`, `PreferWriteOnlyAttributeIf`;
    `objectvalidation.ExactlyOneOfIfObjectSet`;
  - defaults: `defaults.AccountIDRequired()`, `GetDefaultScalrAccountID()`;
  - values (`schema.go`): `SetIfKnownString/Bool/Int`,
    `FlattenRelationshipIDsSet/List`, `ExpandRelationshipIDsSet/List`.
- `MarkdownDescription` on the resource and on every attribute, written for users: what it is,
  allowed values in backticks, default, and the ID format (`env-<RANDOM STRING>`) for IDs.
- `id`: `Computed` + `stringplanmodifier.UseStateForUnknown()`.
- `account_id` (account-scoped objects): `Optional` + `Computed` + `Default: defaults.AccountIDRequired()`
  + `RequiresReplace()`.
- Attributes the API cannot change after create: `RequiresReplace()`.
- Server-computed attributes that do not change on update: `UseStateForUnknown()`; otherwise every
  plan shows `(known after apply)`.
- `Optional` + `Computed` only when the server fills a default the user may override. Pick one
  mode per attribute and keep it; changing modes later is a breaking schema change.
- Validate in the schema (validators, `ConfigValidators`, `ValidateConfig`) so errors appear at
  plan time, not apply.
- Enumerations: `stringvalidator.OneOf(...)` and list the values in the description.
- Secrets: `Sensitive: true`. Never read a masked secret back from the API into state. If the API
  never returns the value, keep the planned value in state on Read and document that drift is
  not detected. For new secrets consider a write-only
  attribute (`WriteOnly: true`, see `value_wo` in `variable_schemas.go`).
- To-many relationships: `types.Set` of IDs, mapped with `framework.FlattenRelationshipIDsSet` /
  `ExpandRelationshipIDsSet`. On update always send the full set; an omitted relationship can be
  cleared by the server ([CONTRIBUTING.md § to-many](../../../CONTRIBUTING.md#working-with-to-many-relationships)).
- Data source attributes mirror the resource's names and descriptions; all are `Computed` except
  the lookup keys.

## CRUD rules

- Create/Update: build the request from the plan. Use `value.Set`, `value.SetPtrMaybe`,
  `framework.SetIfKnown*`. On Update send only what can change; `value.Unset` leaves the server
  value, `value.Null` clears it. Clearing an optional attribute in config must clear it on the
  server.
- After Create/Update, set state from the API response (one `refresh`/`...FromAPI` function shared
  with Read), not from the plan, except for values the API does not return.
- Read: `errors.Is(err, client.ErrNotFound)` → `resp.State.RemoveResource(ctx)` and return.
  Legacy code: `scalr.ErrResourceNotFound` (v1), `d.SetId("")` (SDK).
- Delete: ignore not-found.
- Errors: `resp.Diagnostics.AddError("Error <verb>ing <thing>", err.Error())`; attach to an
  attribute with `AddAttributeError` when the input is at fault.
- Absent optional API values become `null` (`types.StringPointerValue`), never `""`.
- `ImportState`: `resource.ImportStatePassthroughID` for plain IDs; parse and validate composite
  IDs with a clear error message.

## Wire it up

1. Register in `internal/provider/provider.go` `Resources()` / `DataSources()`.
2. Examples: `examples/resources/scalr_<name>/resource.tf` (+ `import.sh` with
   `terraform import scalr_<name>.example <id>`), `examples/data-sources/scalr_<name>/data-source.tf`.
   Use placeholder IDs like `env-xxxxxxxxxx`.
3. Complex pages only: a template in `templates/resources/<name>.md.tmpl`.
4. `go generate` and commit the new `docs/` pages. Expect `position:` front-matter changes in
   other pages: page ordering is regenerated alphabetically.
5. Tests: follow [provider-acceptance-tests](../provider-acceptance-tests/SKILL.md). Minimum for a
   resource: create, update of every updatable attribute, import with `ImportStateVerify`
   (list secrets in `ImportStateVerifyIgnore`), plan-time validation errors, `CheckDestroy`.
   Data source: lookup by each key, not-found error, conflicting keys error.
6. Changelog: `**New resource:** \`scalr_<name>\`` / `**New data source:** \`scalr_<name>\`` under
   `### Added`, per [provider-changelog](../provider-changelog/SKILL.md).

## Done check

- The checks in [AGENTS.md § Agent Workflow](../../../AGENTS.md#agent-workflow) ("Before
  reporting done"). go-scalr was likely bumped, so `go mod tidy -diff` applies, and no `replace`
  may be left in `go.mod`.
- The new acceptance tests run with `make testacc-one` when `.scalr-testenv.env` exists
  ([provider-acceptance-tests](../provider-acceptance-tests/SKILL.md#running-tests-locally));
  otherwise say they were not run and rely on CI.
- Report: files added, API/client version used, fatmouse dependency, tests not run.
