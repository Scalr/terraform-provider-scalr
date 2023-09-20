data "scalr_agent_pool" "example1" {
  id         = "apool-xxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_agent_pool" "example2" {
  name       = "default-pool"
  account_id = "acc-xxxxxxx"
}
