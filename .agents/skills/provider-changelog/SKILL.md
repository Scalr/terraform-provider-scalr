---
name: provider-changelog
description: >-
  MUST be used when writing or reviewing CHANGELOG.md entries for the Scalr
  provider, deciding whether a change needs one, or preparing the changelog for
  a release (turning [Unreleased] into a version). Covers the impact gate,
  sections, entry format and release steps.
---

# Provider changelog

`CHANGELOG.md` follows Keep a Changelog. Readers are Terraform users upgrading the provider. Full
contributor rules: [CONTRIBUTING.md § Updating the changelog](../../../CONTRIBUTING.md#updating-the-changelog).

## Impact gate

Record a change when a user can observe it in configuration, plan, apply, import, state, or docs:
new resource/data source/attribute, changed behaviour or default, new validation, a bug a user can
hit, deprecation, removal, a new minimum Scalr version.

Omit: refactors, SDK → Framework migrations with no behaviour change, tests, CI, go-scalr bumps
with no user effect, docs-only typo fixes. Dependency bumps for security advisories are omitted
unless users must act.

## Sections

Under `## [Unreleased]`, in this order, only those with entries:

- `### Added`: new resources, data sources, attributes, accepted values. A new attribute is
  Added, not Changed.
- `### Changed`: behaviour changes of existing things, including `""` → `null`, new plan-time
  errors, new `RequiresReplace`, supported values extended on an existing attribute.
- `### Fixed`: describe the symptom the user saw.
- `### Deprecated`: name the replacement.
- `### Removed`
- `### Required`: `- scalr-server >= \`X.Y.Z\`` when the provider needs a newer Scalr.

## Entry format

```markdown
- **New resource:** `scalr_datadog_integration` ([#562](https://github.com/Scalr/terraform-provider-scalr/pull/562)).
- **New data source:** `scalr_datadog_integration` ([#562](https://github.com/Scalr/terraform-provider-scalr/pull/562)).
- `scalr_drift_detection`: new attribute `time_window` to set the preferred UTC time window for drift checks ([#565](https://github.com/Scalr/terraform-provider-scalr/pull/565)).
- `data.scalr_workspace`: new attribute `runner_image_version_id` ([#561](https://github.com/Scalr/terraform-provider-scalr/pull/561)).
- `scalr_iam_team`: do not warn about externally managed team membership when `users` is empty.
```

- Start with the type in backticks: `scalr_x` for the resource, `data.scalr_x` for the data
  source. One line per type; a change to both gets two lines.
- Attribute paths in backticks, nested with dots (`scalr_provider_configuration.google`).
- One sentence. Say what the user can do or what they will see, not how it was implemented.
- End with the PR link when the PR number is known (`gh pr view --json number`). Do not invent
  one; leave it off and mention it to the user.
- Plain text, no em dashes; bold only for the `**New resource:**` / `**New data source:**`
  labels.

## Release preparation

When asked to prepare release `X.Y.Z` (a PR into the `release/X.Y.Z` branch, titled
`Update CHANGELOG for X.Y.Z`; the merge is backported to `master` automatically):

1. Rename `## [Unreleased]` to `## [X.Y.Z] - YYYY-MM-DD` and add an empty `## [Unreleased]` above.
2. Choose the version: removals or other breaking changes → major; Added entries → minor;
   only Fixed → patch. Ask if the user's version disagrees.
3. Merge duplicates, order each section resources first then data sources, fix tense and links.
4. Check every entry against the impact gate and against `git log <previous tag>..HEAD`; list
   merged PRs that changed behaviour but have no entry.

## Before finishing

Check: section, type prefix, one line per type, link, wording describes user effect, nothing from
the omit list.
