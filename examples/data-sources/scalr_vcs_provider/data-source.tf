data "scalr_vcs_provider" "example1" {
  id         = "vcs-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_vcs_provider" "example2" {
  name       = "example"
  account_id = "acc-xxxxxxxxxx"
}
