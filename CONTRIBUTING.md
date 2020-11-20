# How to contribute to the Scalr provider
## Updating the changelog
Each PR should update the CHANGELOG.md file to include all changes affecting provider behaviour or compatibility. The version and date heading should be **Unreleased** and updated after being merged and released on the main branch.

The format of the changelog for each release looks like this:


**[1.0.0-rc5] - 2020-09-03** (version and date of release)

**Added**


In this section we need to mention any resources, attributes or other features that have been added. This is generally a list of new things the user can do after this release. Example:

- `scalr_workspace`: new attribute `environment_id` (Scalr environment ID, replaces `organization`) ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))

**Changed**


Existing resources, attributes or behaviour which changes in this release. Attribute changes are likely to break compatibility with older provider configs, however we should never break the state. If any attribute changes in this release it needs to include a state migration and a test for it.

Sometimes there will be changes in go-scalr which may affect the minimum required version of Scalr. In such case we can give a quick overview of changes in go-scalr without going too deep into its internals. Example:

- `scalr_workspace`: attribute `id` is now in the `ws-<RANDOM STRING>` format ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))

**Removed**


This section should detail any resources, attributes or features removed from the provider in this version. This generally means the user will need to make changes to their configs after updating. Example:

- `scalr_workspace`: drop attribute `organization` in favour of `environment_id` ([#11](https://github.com/Scalr/terraform-provider-scalr/pull/11))

**Required**


If the provider relies on changes introduced in a more recent version of Scalr we should specify its version here. If the provider can be used with an older version of Scalr there is no need to update this version from the previous release.

- scalr server >= `8.0.1-beta.20200901`