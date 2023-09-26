data "scalr_environment" "example1" {
  id         = "env-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_environment" "example2" {
  name       = "environment-name"
  account_id = "acc-xxxxxxxxxx"
}
