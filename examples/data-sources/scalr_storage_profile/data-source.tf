data "scalr_storage_profile" "example_sp" {
  name = "my-storage-profile"
}

data "scalr_storage_profile" "default_sp" {
  default = true
}
