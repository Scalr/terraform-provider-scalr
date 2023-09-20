data "scalr_service_account" "example1" {
  id         = "sa-xxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_service_account" "example2" {
  email      = "name@account.scalr.io"
  account_id = "acc-xxxxxxx"
}
