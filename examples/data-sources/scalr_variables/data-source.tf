data "scalr_variables" "example" {
  keys             = ["key1", "key2", "key3"]
  category         = "terraform" # or shell
  account_id       = "acc-tgobtsrgo3rlks8"
  envrironment_ids = ["env-sv0425034857d22", "null"]
  workspace_ids    = ["ws-tlbp7litrs55vgg", "null"]
}
