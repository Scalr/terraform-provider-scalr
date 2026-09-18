resource "scalr_provider_configuration" "aws_module_test" {
  name                      = "aws-module-test"
  is_allowed_in_module_test = true

  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    access_key       = "AKIAXXXXXXXXXXXXXXXX"
    secret_key       = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  }
}

resource "scalr_module_test_configuration" "example" {
  module_id = scalr_module.example.id
  enabled   = true
}

resource "scalr_module_test_provider_configuration_link" "example" {
  test_configuration_id     = scalr_module_test_configuration.example.id
  provider_configuration_id = scalr_provider_configuration.aws_module_test.id
}
