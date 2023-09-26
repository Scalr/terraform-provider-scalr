data "scalr_policy_group" "example1" {
  id         = "pgrp-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_policy_group" "example2" {
  name       = "instance_types"
  account_id = "acc-xxxxxxxxxx"
}
