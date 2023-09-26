data "scalr_endpoint" "example1" {
  id         = "ep-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_endpoint" "example2" {
  name       = "endpoint_name"
  account_id = "acc-xxxxxxxxxx"
}
