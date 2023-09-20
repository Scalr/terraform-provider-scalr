data "scalr_webhook" "example1" {
  id         = "wh-xxxxxxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_webhook" "example2" {
  name       = "webhook_name"
  account_id = "acc-xxxxxxx"
}
