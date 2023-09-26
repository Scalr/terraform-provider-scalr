data "scalr_variable" "example1" {
  id         = "var-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_variable" "example2" {
  key            = "key"
  category       = "terraform"
  account_id     = "acc-xxxxxxxxxx"
  environment_id = "env-xxxxxxxxxx"
  workspace_id   = "ws-xxxxxxxxxx"
}
