# Configure the Scalr Provider
provider "scalr" {
  hostname = var.hostname
  token    = var.token
}

# Create a workspace
resource "scalr_workspace" "example" {
  name            = "my-workspace-name"
  environment_id  = "env-xxxxxxxxx"
  vcs_provider_id = "my_vcs_provider"
  vcs_repo {
    identifier = "org/repo"
    branch     = "dev"
  }
}
