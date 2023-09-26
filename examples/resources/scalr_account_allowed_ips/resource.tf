resource "scalr_account_allowed_ips" "default" {
  account_id  = "acc-xxxxxxxxxx"
  allowed_ips = ["127.0.0.1", "192.168.0.0/24"]
}
