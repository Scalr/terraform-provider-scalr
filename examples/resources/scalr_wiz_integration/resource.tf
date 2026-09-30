resource "scalr_wiz_integration" "example" {
  name          = "wiz"
  client_id     = "svc-account-client-id"
  client_secret = var.wiz_client_secret

  default_policies = ["Default IaC policy"]
  autofail         = false
  endpoint_mode    = "commercial"

  environments = ["env-xxxxxxxxxx"] # or ["*"] for all environments
}
