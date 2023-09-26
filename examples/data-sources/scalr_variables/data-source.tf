data "scalr_variables" "example" {
  keys             = ["key1", "key2", "key3"]
  category         = "terraform" # or shell
  account_id       = "acc-xxxxxxxxxx"
  envrironment_ids = ["env-xxxxxxxxxx", "null"]
  workspace_ids    = ["ws-xxxxxxxxxx", "null"]
}
