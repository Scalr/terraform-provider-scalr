resource "scalr_provider_configuration" "azurerm" {
  name       = "azurerm"
  account_id = "acc-xxxxxxxxx"
  azurerm {
    client_id       = "my-client-id"
    client_secret   = "my-client-secret"
    subscription_id = "my-subscription-id"
    tenant_id       = "my-tenant-id"
  }
}
