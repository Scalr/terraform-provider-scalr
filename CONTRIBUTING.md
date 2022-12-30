# How to contribute to the Scalr provider

## Updating the changelog

Each PR should update the [CHANGELOG.md](./CHANGELOG.md) file to include all changes affecting provider behaviour or compatibility. The version and date heading should be **Unreleased** and updated after being merged and released on the main branch.

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

## Code guidelines

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
- be consistent declaring and initiating variables:
  - `var a int` when declaring the variable
  - `a := 1` walrus notation when declaring and initializing the variable
  - when a slice must be initialized with an empty slice instead of zero value,
    prefer allocating it with `make` function instead of empty slice literal (`make([]int, 0)` over `[]int{}`)

    > **Note**
    > When choosing the initial value for slice, take into account that zero-value slice marshals into `null`,
    while an empty slice will produce `[]`.
