resource "scalr_docker_integration" "example" {
  name               = "ghcr"
  registry_url       = "https://ghcr.io"
  username           = "my-user"
  password           = var.ghcr_token
  export_credentials = true
}
