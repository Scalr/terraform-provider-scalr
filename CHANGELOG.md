# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0-rc26] - 2022-01-21

### Changed
- **New resource:** `scalr_run_triggers` ([#102](https://github.com/Scalr/terraform-provider-scalr/pull/102))
- `data.scalr_environment`: allow to obtain scalr_environment by name ([#101](https://github.com/Scalr/terraform-provider-scalr/pull/101))
- `data.scalr_environment`: `id` become optional  ([#101](https://github.com/Scalr/terraform-provider-scalr/pull/101))
- `data.scalr_environment`: added new optional attribute `name` ([#101](https://github.com/Scalr/terraform-provider-scalr/pull/101))
- `data.scalr_environment`: added new optional attribute `account_id` ([#101](https://github.com/Scalr/terraform-provider-scalr/pull/101))

### Required

- scalr-server >= `8.9.0`

## [1.0.0-rc25] - 2021-11-24

### Changed

- `data.scalr_role`: argument `account_id` is optional now ([#97](https://github.com/Scalr/terraform-provider-scalr/pull/97))

## [1.0.0-rc24] - 2021-11-12

- `data.scalr_webhook`: fixed broken webhook enabled filter ([#93](https://github.com/Scalr/terraform-provider-scalr/pull/93))

## [1.0.0-rc23] - 2021-11-05

- `scalr_workspace`: attribute `vcs_repo.path` has been deprecated ([#92](https://github.com/Scalr/terraform-provider-scalr/pull/92))

### Added

- **New resource:** `scalr_iam_team` ([#96](https://github.com/Scalr/terraform-provider-scalr/pull/96))
- **New data source:** `scalr_iam_team` ([#96](https://github.com/Scalr/terraform-provider-scalr/pull/96))
- **New data source:** `scalr_iam_user` ([#96](https://github.com/Scalr/terraform-provider-scalr/pull/96))
- **New resource:** `scalr_policy_group` ([#94](https://github.com/Scalr/terraform-provider-scalr/pull/94))
- **New resource:** `scalr_policy_group_linkage` ([#94](https://github.com/Scalr/terraform-provider-scalr/pull/94))
- **New data source:** `scalr_policy_group` ([#94](https://github.com/Scalr/terraform-provider-scalr/pull/94))

### Required

- scalr-server >= `8.3.0`

## [1.0.0-rc22] - 2021-10-22

### Added

- **New resource:** `scalr_agent_pool` ([#85](https://github.com/Scalr/terraform-provider-scalr/pull/85))
- **New data source:** `scalr_agent_pool` ([#85](https://github.com/Scalr/terraform-provider-scalr/pull/85))
- **New resource:** `scalr_agent_pool_token` ([#85](https://github.com/Scalr/terraform-provider-scalr/pull/85))
- `scalr_workspace`: added new attribute `agent_pool_id` ([#85](https://github.com/Scalr/terraform-provider-scalr/pull/85))
- **New resource:** `scalr_vcs_provider` ([#88](https://github.com/Scalr/terraform-provider-scalr/pull/88))
- **New data source:** `scalr_vcs_provider` ([#89](https://github.com/Scalr/terraform-provider-scalr/pull/89))

### Required

- scalr-server >= `8.1.0`

## [1.0.0-rc21] - 2021-10-01

### Added

- **New data source:** `scalr_module_version` ([#76](https://github.com/Scalr/terraform-provider-scalr/pull/76))
- **New resource:** `scalr_module` ([#76](https://github.com/Scalr/terraform-provider-scalr/pull/76))

### Changed

- `scalr_workspace`: new attribute `module_version_id` ([#76](https://github.com/Scalr/terraform-provider-scalr/pull/76))

### Fixed

- panic when retrying failed request ([#87](https://github.com/Scalr/terraform-provider-scalr/pull/87))
- `data.scalr_access_policy`: return error if access policy is not found ([#83](https://github.com/Scalr/terraform-provider-scalr/pull/83))
- `data.scalr_environment`: return error if environment is not found ([#83](https://github.com/Scalr/terraform-provider-scalr/pull/83))
- `scalr_environment`: fixed crash while reading environment without proper permissions ([#82](https://github.com/Scalr/terraform-provider-scalr/pull/82))

### Required

- scalr-server >= `8.0.1-beta.20210930`

## [1.0.0-rc20] - 2021-09-10

### Fixed

- `scalr_environment`: fixed handling of empty strings in `cloud_credentials` and `policy_groups` attributes ([#81](https://github.com/Scalr/terraform-provider-scalr/pull/81))
- `scalr_webhook`: fixed handling of empty strings in `events` attribute ([#81](https://github.com/Scalr/terraform-provider-scalr/pull/81))
- `scalr_access_policy`: fixed handling of empty strings in `role_ids` attribute ([#81](https://github.com/Scalr/terraform-provider-scalr/pull/81))
- `scalr_role`: fixed handling of empty strings in `permissions` attribute ([#81](https://github.com/Scalr/terraform-provider-scalr/pull/81))
- `scalr_workspace`: fixed handling of empty strings in `vcs_repo.trigger_prefixes` attribute ([#81](https://github.com/Scalr/terraform-provider-scalr/pull/81))

### Required

- scalr server >= `8.0.1-beta.20210810`

## [1.0.0-rc19] - 2021-08-19

### Added

- **New data source:** `scalr_role` ([#69](https://github.com/Scalr/terraform-provider-scalr/pull/69))
- **New data source:** `scalr_access_policy` ([#69](https://github.com/Scalr/terraform-provider-scalr/pull/69))
- **New resource:** `scalr_role` ([#69](https://github.com/Scalr/terraform-provider-scalr/pull/69))
- **New resource:** `scalr_access_policy` ([#69](https://github.com/Scalr/terraform-provider-scalr/pull/69))

### Changed

- `scalr_variable`: new attribute `description` ([#73](https://github.com/Scalr/terraform-provider-scalr/pull/73))
- `scalr_workspace`: added new attribute `has_resources` ([#63](https://github.com/Scalr/terraform-provider-scalr/pull/63))
- `data.scalr_workspace`: added new attribute `has_resources` ([#63](https://github.com/Scalr/terraform-provider-scalr/pull/63))
- `scalr_workspace`: added new attribute `vcs_repo.dry_runs_enabled` ([#70](https://github.com/Scalr/terraform-provider-scalr/pull/70))
- `data.scalr_workspace`: added new attribute `vcs_repo.dry_runs_enabled` ([#70](https://github.com/Scalr/terraform-provider-scalr/pull/70))

### Fixed

 - `scalr_environment`: fix unlinking cloud credentials ([#71](https://github.com/Scalr/terraform-provider-scalr/pull/71))
 - `scalr_workspace`: fix removing hooks if it removed from template ([#72](https://github.com/Scalr/terraform-provider-scalr/pull/72))

### Required

- scalr server >= `8.0.1-beta.20210810`

## [1.0.0-rc18] - 2021-07-22

### Changed

- `scalr_workspace`: make `working_directory` attribute non-computable, set default value to `""` ([#66](https://github.com/Scalr/terraform-provider-scalr/pull/66))

### Fixed

- `scalr_variable`: fix inability to create sensitive variable ([#68](https://github.com/Scalr/terraform-provider-scalr/pull/68))
- `scalr_workspace`: fix error changing working directory of a workspace to empty: plan outputs to empty diff ([#66](https://github.com/Scalr/terraform-provider-scalr/pull/66))

### Required

- scalr server >= `8.0.1-beta.20210407`

## [1.0.0-rc17] - 2021-07-08

### Added

- `scalr_workspace`: new attribute `hooks` ([#65](https://github.com/Scalr/terraform-provider-scalr/pull/65))
- `data.scalr_workspace`: new attribute `hooks` ([#65](https://github.com/Scalr/terraform-provider-scalr/pull/65))

### Changed

- `scalr_variable`: new attribute value `shell` for `scalr_variable.category` in order to create shell variable.
`env` category value is deprecated. ([#59](https://github.com/Scalr/terraform-provider-scalr/pull/64))


## [1.0.0-rc16] - 2021-05-25

### Changed

- `scalr_variable`: make `environment_id`, `workspace_id` and `account_id` attributes computable ([#60](https://github.com/Scalr/terraform-provider-scalr/pull/62))

### Fixed

- Error changing scope for variable `var-<id>`: scope is immutable attribute

### Required

- scalr server >= `8.0.1-beta.20210407`

## [1.0.0-rc15] - 2021-04-22

### Added

 - The version number in terraform provider binary name and in User-Agent header during API calls to Scalr server ([#60](https://github.com/Scalr/terraform-provider-scalr/pull/60))
 - `scalr_workspace`: new attribute `vcs_repo.path` ([#59](https://github.com/Scalr/terraform-provider-scalr/pull/59))
 - `scalr_workspace`: new attribute `vcs_repo.trigger_prefixes` ([#59](https://github.com/Scalr/terraform-provider-scalr/pull/59))

### Changed

- `scalr_variable`: variable's scope becomes immutable (can not change `workspace_id`, `environment_id` or `account_id`) ([#57](https://github.com/Scalr/terraform-provider-scalr/pull/57))
- `scalr_variable`: can not change `key` attribute for sensitive variable ([#57](https://github.com/Scalr/terraform-provider-scalr/pull/57))
- `scalr_endpoint`: refresh state after manually endpoint deleting ([#55](https://github.com/Scalr/terraform-provider-scalr/pull/55))

### Required

- scalr server >= `8.0.1-beta.20210407`

## [1.0.0-rc14] - 2021-03-11

### Added

- `scalr_variable`: new attribute `final` ([#50](https://github.com/Scalr/terraform-provider-scalr/pull/50))
- `scalr_variable`: new attribute `force` ([#50](https://github.com/Scalr/terraform-provider-scalr/pull/50))
- `scalr_variable`: new attribute `environment_id` ([#50](https://github.com/Scalr/terraform-provider-scalr/pull/50))
- `scalr_variable`: new attribute `account_id` ([#50](https://github.com/Scalr/terraform-provider-scalr/pull/50))

### Changed

- `scalr_variable`: attribute `workspace_id` is optional ([#50](https://github.com/Scalr/terraform-provider-scalr/pull/50))

### Required

- scalr server >= `8.0.1-beta.20201202`

## [1.0.0-rc13] - 2021-02-04

### Changed

- Fix inconsistency in migration (since 1.0.0-rc5) ([#49](https://github.com/Scalr/terraform-provider-scalr/pull/49))

### Required

- scalr server >= `8.0.1-beta.20201202`

## [1.0.0-rc12] - 2021-01-28

### Changed

- Fix for workspace vcs_repo state migration panic ([#46](https://github.com/Scalr/terraform-provider-scalr/pull/46))

### Required

- scalr server >= `8.0.1-beta.20201202`

## [1.0.0-rc11] - 2020-12-10

### Changed

- `data.scalr_current_run` use the `SCALR_RUN_ID` environment variable to read the current run ID ([#42](https://github.com/Scalr/terraform-provider-scalr/pull/42))

### Required

- scalr server >= `8.0.1-beta.20201202`

## [1.0.0-rc10] - 2020-12-03

### Changed

- `scalr_webhook` attribute `enabled` is optional with default: `true`. ([#40](https://github.com/Scalr/terraform-provider-scalr/pull/40))
- `scalr_endpoint` attribute `secret_key` is optional and sensitive.

### Removed

- `scalr_workspace`: drop attribute `queue_all_runs`. ([#40](https://github.com/Scalr/terraform-provider-scalr/pull/40))
- `scalr_endpoint`: drop attribute `http_method`. ([#40](https://github.com/Scalr/terraform-provider-scalr/pull/40))

### Required

- scalr server >= `8.0.1-beta.20201125`

## [1.0.0-rc9] - 2020-11-12

### Added
- **New data source:** `scalr_environment` ([#10](https://github.com/Scalr/terraform-provider-scalr/pull/34))
- **New resource:** `scalr_environment` ([#10](https://github.com/Scalr/terraform-provider-scalr/pull/34))

### Required

- scalr server >= `8.0.1-beta.20201104`

## [1.0.0-rc8] - 2020-10-29

### Changed

- `data.scalr_current_run` use VCS revision API instead of ingress attributes API to get VCS revision data ([#24](https://github.com/Scalr/terraform-provider-scalr/pull/24))

### Required

- scalr server >= `8.0.1-beta.20201019`

## [1.0.0-rc7] - 2020-10-01

### Added
- **New data source:** `scalr_endpoint` ([#10](https://github.com/Scalr/terraform-provider-scalr/pull/10))
- **New data source:** `scalr_webhook` ([#10](https://github.com/Scalr/terraform-provider-scalr/pull/10))
- **New resource:** `scalr_endpoint` ([#10](https://github.com/Scalr/terraform-provider-scalr/pull/10))
- **New resource:** `scalr_webhook` ([#10](https://github.com/Scalr/terraform-provider-scalr/pull/10))

### Required

- scalr server >= `8.0.1-beta.20200917`

## [1.0.0-rc6] - 2020-09-10

### Added

- `scalr_workspace`: new attribute `vcs_provider_id` (Scalr vcs provider ID, replaces `vcs_repo.oauth_token_id`)  ([#17](https://github.com/Scalr/terraform-provider-scalr/pull/17))

### Removed

- `scalr_workspace`: drop attribute `vcs_repo.oauth_token_id` ([#17](https://github.com/Scalr/terraform-provider-scalr/pull/17))

### Required

- scalr server >= `8.0.1-beta.20200903`

## [1.0.0-rc5] - 2020-09-03

### Added

- `scalr_workspace`: new attribute `environment_id` (Scalr environment ID, replaces `organization`) ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))
- `provider`: new environment variable `SCALR_HOSTNAME` (Scalr hostname, replaces `TFE_HOSTNAME`) ([#12](https://github.com/Scalr/terraform-provider-scalr/pull/12))
- `provider`: new environment variable `SCALR_TOKEN` (Scalr token, replaces `SCALR_TOKEN`) ([#12](https://github.com/Scalr/terraform-provider-scalr/pull/12))

### Changed

- `scalr_workspace`: attribute `id` is now in the `ws-<RANDOM STRING>` format ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))

### Removed

- `scalr_workspace`: drop attribute `organization` in favour of `environment_id` ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))
- `scalr_workspace`: drop attribute `external_id` in favour of `id` ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))
- `scalr_workspace`: drop attribute `vcs_repo.ingress_submodules` ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))
- `provider`: drop environment variable `TFE_HOSTNAME` in favour of `SCALR_HOSTNAME` ([#12](https://github.com/Scalr/terraform-provider-scalr/pull/12))
- `provider`: drop environment variable `TFE_TOKEN` in favour of `SCALR_TOKEN` ([#12](https://github.com/Scalr/terraform-provider-scalr/pull/12))

### Required

- scalr server >= `8.0.1-beta.20200901`

## [1.0.0-rc3] - 2020-07-30

### Added

- `scalr_workspace`: new attribute `vcs_repo.path` ([#8](https://github.com/Scalr/terraform-provider-scalr/pull/8))
- `data.scalr_workspace`: new attribute `vcs_repo.path` ([#8](https://github.com/Scalr/terraform-provider-scalr/pull/8))

### Changed

- `data.scalr_current_run` will return empty values on the local operation backend instead of an error ([#9](https://github.com/Scalr/terraform-provider-scalr/pull/9))

### Required

- scalr server >= `8.0.1-beta.20200709`

## [1.0.0-rc2] - 2020-07-10

### Added

- `data.scalr_current_run`: new attribute `environment_id` ([#6](https://github.com/Scalr/terraform-provider-scalr/pull/6))
- `data.scalr_current_run`: new attribute `workspace_name` ([#6](https://github.com/Scalr/terraform-provider-scalr/pull/6))

### Removed

- `data.scalr_current_run`: drop attribute `workspace` ([#6](https://github.com/Scalr/terraform-provider-scalr/pull/6))

## [1.0.0-rc1] - 2020-07-01

Requires Scalr 8.0.1-beta.20200625 at least

### Added

- **New data source:** `scalr_current_run` ([#2](https://github.com/Scalr/terraform-provider-scalr/pull/2))
- `data.scalr_workspace`: new attribute `created_by` ([#5](https://github.com/Scalr/terraform-provider-scalr/pull/5))
- `scalr_workspace`: new attribute `created_by` ([#5](https://github.com/Scalr/terraform-provider-scalr/pull/5))

## [0.0.0-rc2] - 2020-03-30

### Added

- **New data source:** `scalr_workspace` ([#1](https://github.com/Scalr/terraform-provider-scalr/pull/1))
- **New data source:** `scalr_workspace_ids` ([#1](https://github.com/Scalr/terraform-provider-scalr/pull/1))
- **New resource:** `scalr_variable` ([#1](https://github.com/Scalr/terraform-provider-scalr/pull/1))
- **New resource:** `scalr_workspace` ([#1](https://github.com/Scalr/terraform-provider-scalr/pull/1))

## [0.0.0-rc1] - 2020-03-25

### Added

- Initial release.

[Unreleased]: https://github.com/Scalr/terraform-provider-scalr/compare/v1.0.0-rc26...HEAD
[1.0.0-rc26]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc26
[1.0.0-rc25]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc25
[1.0.0-rc24]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc24
[1.0.0-rc23]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc23
[1.0.0-rc22]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc22
[1.0.0-rc21]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc21
[1.0.0-rc20]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc20
[1.0.0-rc19]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc19
[1.0.0-rc18]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc18
[1.0.0-rc17]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc17
[1.0.0-rc16]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc16
[1.0.0-rc15]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc15
[1.0.0-rc14]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc14
[1.0.0-rc13]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc13
[1.0.0-rc12]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc12
[1.0.0-rc11]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc11
[1.0.0-rc10]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc10
[1.0.0-rc9]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc9
[1.0.0-rc8]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc8
[1.0.0-rc7]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc7
[1.0.0-rc6]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc6
[1.0.0-rc5]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc5
[1.0.0-rc3]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc3
[1.0.0-rc2]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc2
[1.0.0-rc1]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v1.0.0-rc1
[0.0.0-rc2]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v0.0.0-rc2
[0.0.0-rc1]: https://github.com/Scalr/terraform-provider-scalr/releases/tag/v0.0.0-rc1
