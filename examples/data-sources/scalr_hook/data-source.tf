data "scalr_hook" "example1" {
  id         = "hook-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_hook" "example2" {
  name       = "production"
  account_id = "acc-xxxxxxxxxx"
}
