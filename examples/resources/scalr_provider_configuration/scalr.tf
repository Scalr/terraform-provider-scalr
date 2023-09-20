resource "scalr_provider_configuration" "scalr" {
  name       = "scalr"
  account_id = "acc-xxxxxxxxx"
  scalr {
    hostname = "scalr.host.example.com"
    token    = "my-scalr-token"
  }
}
