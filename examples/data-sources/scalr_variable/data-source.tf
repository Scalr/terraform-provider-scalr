data "scalr_variable" "example1" {
  id         = "var-xxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_variable" "example2" {
  key            = "key"
  category       = "terraform"
  account_id     = "acc-xxxxxxx"
  environment_id = "env-xxxxxxx"
  workspace_id   = "ws-xxxxxxx"
}
