resource "scalr_provider_configuration" "aws" {
  name                   = "aws_dev_us_east_1"
  account_id             = "acc-xxxxxxxxx"
  export_shell_variables = false
  environments           = ["env-yyyyyyyyy"]
  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    secret_key       = "my-secret-key"
    access_key       = "my-access-key"
  }
}
