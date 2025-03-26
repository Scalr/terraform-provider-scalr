resource "scalr_environment" "test" {
  name                            = "test-env"
  account_id                      = "acc-xxxxxxxxxx"
  policy_groups                   = ["pgrp-xxxxxxxxxx", "pgrp-yyyyyyyyyy"]
  default_provider_configurations = ["pcfg-xxxxxxxxxx", "pcfg-yyyyyyyyyy"]
}
