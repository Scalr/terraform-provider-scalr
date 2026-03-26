resource "scalr_module" "example" {
  namespace_id    = scalr_module_namespace.shared.id
  vcs_provider_id = "vcs-xxxxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
    path       = "example/terraform-<provider>-<name>"
    tag_prefix = "aws/"
  }
}

# Edge case: VCS path does not follow terraform-<provider>-<name> (e.g. extra hyphens).
# Set module_provider and name so the registry maps the module correctly.
resource "scalr_module" "example_explicit_provider_and_name" {
  namespace_id    = scalr_module_namespace.shared.id
  vcs_provider_id = "vcs-xxxxxxxxxx"

  module_provider = "couchbasecapella"
  name            = "infra"

  vcs_repo {
    identifier = "org/repo"
    path       = "example/terraform-couchbase-capella-infra"
  }
}
