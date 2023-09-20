resource "scalr_provider_configuration" "oidc" {
  name                   = "oidc_dev_us_east_1"
  account_id             = "acc-xxxxxxxxx"
  export_shell_variables = false
  environments           = ["*"]
  aws {
    credentials_type = "oidc"
    role_arn         = "arn:aws:iam::123456789012:role/scalr-oidc-role"
    audience         = "aws.scalr-run-workload"
  }
}
