data "scalr_endpoint" "example1" {
  id         = "ep-xxxxxxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_endpoint" "example2" {
  name       = "endpoint_name"
  account_id = "acc-xxxxxxx"
}
