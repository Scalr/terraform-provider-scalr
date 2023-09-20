data "scalr_provider_configuration" "aws" {
  id = "pcfg-xxxxxxx"
}

data "scalr_provider_configuration" "aws_dev" {
  name = "aws_dev"
}

data "scalr_provider_configuration" "azure" {
  provider_name = "azurerm"
}
