data "scalr_container_image" "runner" {
  name       = "scalr/runner"
  visibility = "system"
}

data "scalr_container_image_version" "runner" {
  container_image_id = data.scalr_container_image.runner.id
  version            = "0.5.0"
}

resource "scalr_workspace" "example" {
  name                    = "my-workspace-name"
  environment_id          = "env-xxxxxxxxxx"
  runner_image_version_id = data.scalr_container_image_version.runner.id
}
