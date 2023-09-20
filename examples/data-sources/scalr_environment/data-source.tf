data "scalr_environment" "example1" {
  id         = "env-xxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_environment" "example2" {
  name       = "environment-name"
  account_id = "acc-xxxxxxx"
}
