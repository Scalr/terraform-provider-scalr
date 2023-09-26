resource "scalr_variable" "example1" {
  key          = "my_key_name"
  value        = "my_value_name"
  category     = "terraform"
  description  = "variable description"
  workspace_id = "ws-xxxxxxxxxx"
}

resource "scalr_variable" "example2" {
  key          = "xyz"
  value        = jsonencode(["foo", "bar"])
  hcl          = true
  category     = "terraform"
  workspace_id = "ws-yyyyyyyyyy"
}
