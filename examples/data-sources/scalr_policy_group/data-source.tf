data "scalr_policy_group" "example1" {
  id         = "pgrp-xxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_policy_group" "example2" {
  name       = "instance_types"
  account_id = "acc-xxxxxxx"
}
