data "scalr_vcs_provider" "example1" {
  id         = "vcs-xxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_vcs_provider" "example2" {
  name       = "example"
  account_id = "acc-xxx"
}
