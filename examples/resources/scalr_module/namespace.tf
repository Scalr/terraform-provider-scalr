resource "scalr_module_namespace" "shared" {
  name      = "shared"
  is_shared = true
}

resource "scalr_module_namespace" "private" {
  name         = "private"
  is_shared    = false
  environments = [scalr_environment.example.id]
}

resource "scalr_module" "example_with_namespace" {
  namespace_id    = scalr_module_namespace.shared.id
  vcs_provider_id = "vcs-xxxxxxxxxx"
  vcs_repo {
    identifier = "org/repo"
  }
} 