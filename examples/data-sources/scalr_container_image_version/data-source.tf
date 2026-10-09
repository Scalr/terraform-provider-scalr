data "scalr_container_image" "runner" {
  name       = "scalr/runner"
  visibility = "system"
}

# The latest version of the image
data "scalr_container_image_version" "latest" {
  container_image_id = data.scalr_container_image.runner.id
}

# A specific version of the image
data "scalr_container_image_version" "pinned" {
  container_image_id = data.scalr_container_image.runner.id
  version            = "0.5.0"
}
