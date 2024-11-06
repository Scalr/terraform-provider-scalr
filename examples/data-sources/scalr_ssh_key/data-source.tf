data "scalr_webhook" "example1" {
  id         = "wh-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_webhook" "example2" {
  name       = "webhook_name"
  account_id = "acc-xxxxxxxxxx"
}
