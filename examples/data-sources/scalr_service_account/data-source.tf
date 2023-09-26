data "scalr_service_account" "example1" {
  id         = "sa-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_service_account" "example2" {
  email      = "name@account.scalr.io"
  account_id = "acc-xxxxxxxxxx"
}
