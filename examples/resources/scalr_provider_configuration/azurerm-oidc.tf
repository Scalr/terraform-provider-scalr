resource "scalr_provider_configuration" "azurerm_oidc" {
  name       = "azurerm"
  account_id = "acc-xxxxxxxxx"
  azurerm {
    auth_type       = "oidc"
    audience        = "scalr-workload-identity"
    client_id       = "my-client-id"
    tenant_id       = "my-tenant-id"
    subscription_id = "my-subscription-id"
  }
}
