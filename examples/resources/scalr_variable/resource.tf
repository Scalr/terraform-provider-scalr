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

# Using write-only value (Terraform 1.11+)
resource "scalr_variable" "example3" {
  key              = "secret_key"
  value_wo         = ephemeral.aws_secretsmanager_secret.my_secret.secret_string
  value_wo_version = 1 # Increment to trigger an update when the secret changes
  category         = "terraform"
  sensitive        = true
  workspace_id     = "ws-zzzzzzzzzz"
}
