data "scalr_container_image" "example1" {
  id = "cimg-xxxxxxxxxx"
}

data "scalr_container_image" "example2" {
  name       = "scalr/runner"
  visibility = "system"
}
