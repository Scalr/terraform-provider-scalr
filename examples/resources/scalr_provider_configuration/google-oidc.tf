resource "scalr_provider_configuration" "google" {
  name       = "google_main"
  account_id = "acc-xxxxxxxxxx"
  google {
    auth_type              = "oidc"
    project                = "my-project"
    service_account_email  = "user@example.com"
    workload_provider_name = "projects/123/locations/global/workloadIdentityPools/pool-name/providers/provider-name"
  }
}
