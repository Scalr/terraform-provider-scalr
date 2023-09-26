resource "scalr_environment" "test" {
  name                            = "test-env"
  account_id                      = "acc-xxxxxxxxxx"
  cost_estimation_enabled         = true
  policy_groups                   = ["pgrp-xxxxxxxxxx", "pgrp-yyyyyyyyyy"]
  default_provider_configurations = ["pcfg-xxxxxxxxxx", "pcfg-yyyyyyyyyy"]
}
