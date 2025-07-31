resource "scalr_module" "example" {
  namespace_id    = scalr_module_namespace.shared.id
  vcs_provider_id = "vcs-xxxxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
    path       = "example/terraform-<provider>-<name>"
    tag_prefix = "aws/"
  }
}
