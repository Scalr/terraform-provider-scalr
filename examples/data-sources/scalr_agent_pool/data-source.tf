data "scalr_agent_pool" "example1" {
  id         = "apool-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_agent_pool" "example2" {
  name       = "default-pool"
  account_id = "acc-xxxxxxxxxx"
}
