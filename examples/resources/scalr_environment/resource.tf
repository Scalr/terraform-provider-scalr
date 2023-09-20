resource "scalr_environment" "test" {
  name                            = "test-env"
  account_id                      = "acc-<id>"
  cost_estimation_enabled         = true
  policy_groups                   = ["pgrp-xxxxx", "pgrp-yyyyy"]
  default_provider_configurations = ["pcfg-xxxxx", "pcfg-yyyyy"]
}
