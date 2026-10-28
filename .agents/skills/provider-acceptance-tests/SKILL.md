---
name: provider-acceptance-tests
description: >-
  MUST be used when writing, changing, running or debugging Scalr provider
  tests: acceptance tests (resource.Test, TF_ACC), state upgrader unit tests,
  SDK-upgrade tests, or a failing acc-tests CI job. Covers the CI preview
  environment, seed data, test structure, naming and common flakiness causes.
---

# Provider tests

## Where tests run

- `make test` runs every test without `TF_ACC`; `resource.Test` skips itself, so only unit tests
  execute (30 s timeout per package batch).
- CI `acc-tests` (`.github/workflows/default.yml`) creates a fatmouse preview environment from the
  fatmouse branch with the same name as this branch (fallback `master`), then runs
  `make testacc` (all acceptance tests, `-race`, 45 min) with `SCALR_HOSTNAME`/`SCALR_TOKEN` and
  cloud credentials (`TEST_AWS_*`, `TEST_ARM_*`, ...) from secrets.
- A fatmouse `master` change can break provider tests on any branch that has no same-named
  fatmouse branch.
- Tests rely on that environment's seed data: `defaultAccount` (`acc-svrcncgh453bi8g`),
  `testUser`, `readOnlyRole`, `userRole` in `testing.go`. A Scalr without that seed data fails
  in odd ways. Do not hard-code other IDs; create what the test needs.

## Running tests locally

Only against a disposable test Scalr with the seed data, never a customer or production account.
The developer keeps its settings in the git-ignored `.scalr-testenv.env` (template:
`.scalr-testenv.env.sample`): `SCALR_HOSTNAME`, `SCALR_TOKEN`, and optional credentials for
tests that need them.

```bash
make testacc-one RUN='^TestAccScalrTag_basic$'
make testacc-one RUN='^TestAccScalrTag_'          # every test of one resource
```

The target loads the file, refuses to run without `RUN` or against a `*.scalr.io` host, and
runs `go test` with `TF_ACC=1 -count=1`.

- `.scalr-testenv.env` present: run the tests for the code you changed through the target. Each
  run creates and destroys real objects; never run the whole suite.
- Before the first run, check the file is ignored and untracked
  (`git check-ignore -q .scalr-testenv.env && ! git ls-files --error-unmatch .scalr-testenv.env`).
  Stop if not.
- Never read, print or grep the file, and never export its values in another way. The target is
  the only consumer.
- File missing: do not ask for a host or token in the chat. Tell the developer they can create
  the file from the sample, and report the acceptance tests as not run.

Tests that need external credentials skip themselves when the env vars are missing
(`TEST_DOCKER_REGISTRY_*`, `TEST_INFRACOST_API_KEY`, `githubToken` (literal lower-case env var
name, not set in CI), ...). A skip is not a pass: report it.

## Structure

```go
func TestAccScalrThing_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-thing")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrThingDestroy,
		Steps: []resource.TestStep{
			{Config: testAccScalrThingConfig(name), Check: resource.ComposeTestCheckFunc(...)},
			{ResourceName: "scalr_thing.test", ImportState: true, ImportStateVerify: true},
		},
	})
}
```

- Always `protoV5ProviderFactories(t)` (serves both Framework and SDK resources). Never the old
  SDK-only factories.
- Unique names: `acctest.RandomWithPrefix(...)` or `GetRandomInteger()`. Environments and other
  account-wide objects collide across parallel CI runs otherwise.
- Create the parent objects in the config (an environment for a workspace, etc.) instead of
  relying on existing ones.
- Imports: `github.com/hashicorp/terraform-plugin-testing/...` (`helper/resource`, `plancheck`,
  `terraform`) and go-scalr v2 only. Tests of new resources never import
  `terraform-plugin-sdk/v2` or go-scalr v1 (`github.com/scalr/go-scalr`); many existing tests do,
  do not copy them.
- `CheckDestroy` uses a real client (`createScalrClientV2()`; `createScalrClient()` is v1) and
  expects not-found for every resource of the type in state.
- Config builders are `func testAccScalrThingConfig(...) string` with `fmt.Sprintf`, kept at the
  bottom of the test file.

## What to cover

- Create with required attributes only, then with all attributes.
- Update every updatable attribute, then remove optional ones from config and assert the result
  (`TestCheckNoResourceAttr` or the server default).
- An attribute with `RequiresReplace`: assert replacement with
  `ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(name, plancheck.ResourceActionReplace)}}`.
- Plan stability: a step re-applying the same config with `plancheck.ExpectEmptyPlan()` when
  computed or normalised values are involved.
- Plan-time validation: `PlanOnly: true` + `ExpectError: regexp.MustCompile(...)`. Match a stable
  fragment of the message, not the whole text.
- Import with `ImportStateVerify`; secrets the API does not return go into
  `ImportStateVerifyIgnore`.
- Data sources: each lookup key, not-found, mutually exclusive keys.
- Bug fixes: a test named `TestAcc<Resource>_SCALRCORE_<n>` (or a step in an existing test) that
  fails on the unfixed code.
- SDK upgrades and schema version bumps: `ExternalProviders` step then current provider with
  `ExpectEmptyPlan` (`TestAccScalrTag_UpgradeFromSDK`).

Prefer adding steps to an existing test function over new functions; each `resource.Test` creates
and destroys real objects and adds minutes to CI.

## Unit tests

- State upgraders (SDK): build the old raw state map, call the upgrader, `assertCorrectState`.
- Framework upgraders and helpers: plain table tests.
- `testScalrClient(t)` gives a v1 client with `client.NewMockWorkspaces()` /
  `NewMockVariables()` for code that needs a client without a network.

## Debugging a failing acceptance test

1. Read the CI log for the first failing test and its step number. Ignore the noise of tests
   that failed after it during cleanup.
2. Provider errors (`After applying this test step, the plan was not empty`, `Provider produced
   inconsistent result after apply`, `ImportStateVerify attributes not equivalent`): see the
   symptom table in [provider-fix-issue](../provider-fix-issue/SKILL.md#symptom--where-to-look).
   A secret the API does not return belongs in `ImportStateVerifyIgnore`.
3. 422/403 from the API → check whether the test needs an API change from a fatmouse branch that
   does not exist under the same name, or seed data that changed.
4. Flaky only in CI → name collisions or account-wide limits shared with parallel runs.
