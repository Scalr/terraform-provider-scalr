resource "scalr_environment" "test" {
  name                            = "test-env"
  account_id                      = "acc-xxxxxxxxxx"
  default_provider_configurations = ["pcfg-xxxxxxxxxx", "pcfg-yyyyyyyyyy"]
}
