---
name: provider-fix-issue
description: >-
  MUST be used when investigating or fixing a Scalr provider bug from a Jira
  ticket, GitHub issue, support report, Terraform error output, perpetual diff,
  "inconsistent result after apply", import failure, or crash. Diagnose the
  exact cause before patching, reproduce with a failing test, make the smallest
  fix.
---

# Fix a provider bug

## Workflow

1. Intake. Jira key → fetch the issue (Atlassian MCP when available). Collect the provider
   version, Terraform/OpenTofu version, the config snippet, the exact error or plan output, and
   whether it worked in an earlier provider version. Missing config or error text that the
   diagnosis depends on: ask.
2. Classify the symptom (table below) to pick where to look first.
3. Localize to the resource, the CRUD method or plan modifier, and the attribute.
4. Check the server side. The provider may be right and the API wrong or changed. Compare the
   go-scalr v2 schema with the fatmouse API (`../fatmouse/taco/openapi/openapi-public.yml`, the
   resource code under `../fatmouse/taco/app/`). If the defect is in Scalr, stop and report it as
   a fatmouse issue.
5. Diagnosis gate. Name: the user-visible failure, the code path, the violated assumption, and
   the evidence. If any is missing, report hypotheses and what would decide between them. Do not
   patch on a guess.
6. Reproduce: an acceptance test step `TestAcc<Resource>_SCALRCORE_<n>` (or a step added to an
   existing test) that fails on the current code for the diagnosed reason. Unit test when the bug
   is in pure logic (state upgrader, helper). If you cannot run acceptance tests, write the test
   anyway and say it was not run.
7. Legacy resource: before patching, ask whether to migrate it in this task
   ([AGENTS.md § Legacy resources](../../../AGENTS.md#legacy-resources)). Narrow is the usual
   answer for a small or urgent fix. On a migration, the regression test from step 6 must pass
   against the migrated version.
   Fix the cause with the smallest change. Follow `provider-schema-change` if the fix touches the
   schema or state shape.
8. Verify: the new test, the resource's other tests where possible, and the checks in
   [AGENTS.md § Agent Workflow](../../../AGENTS.md#agent-workflow) ("Before reporting done").
9. Changelog `### Fixed` entry describing the symptom ([provider-changelog](../provider-changelog/SKILL.md)).

## Symptom → where to look

| Symptom | Usual cause |
|---|---|
| Perpetual diff / non-empty plan after apply | Read returns a value the config cannot produce: `""` vs `null`, server default on a non-`Computed` attribute, list where order is not stable (should be a set), normalised value (case, trailing slash, JSON formatting) |
| `Provider produced inconsistent result after apply` | Create/Update writes a state value different from the plan; unknown left in state; computed value not marked `Computed` |
| `(known after apply)` on every plan | Missing `UseStateForUnknown()` on a stable computed attribute |
| Attribute silently cleared on update | Update omits a to-many relationship or sends `value.Null` instead of `value.Unset` (or the reverse: cannot clear because it sends `Unset`) |
| Import fails or `ImportStateVerify` mismatch | Read does not populate the attribute, or import ID parsing |
| Resource deleted outside Terraform causes an error | Read does not treat not-found as gone |
| Error after provider upgrade on existing state | Missing/incorrect state upgrader or schema version, SDK → Framework difference |
| Validation passes plan, fails at apply | Validation in CRUD instead of schema/`ValidateConfig`; unknown values not handled in `ValidateConfig` |
| Panic | Nil pointer on an optional API relationship or attribute (`*string`, `Relationships.X == nil`) |

## Rules

- The trace or error shows where it surfaced; find why the state was reached.
- Do not widen a timeout, add a retry, or suppress a diff with a plan modifier to hide a wrong
  Read. Fix what Read returns.
- Preserve behaviour outside the failing scenario; existing users depend on it.
- Do not commit unless asked.

## Report

- Source and context used.
- Root cause: failure, code path, assumption, evidence. Provider or server defect.
- Fix: files changed, what the user sees differently, what stays the same.
- Verification: commands run and results; acceptance tests not run.
- Risks: state impact, other resources with the same pattern (name them, do not fix them).
