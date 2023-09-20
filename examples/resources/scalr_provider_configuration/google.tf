resource "scalr_provider_configuration" "google" {
  name       = "google_main"
  account_id = "acc-xxxxxxxxx"
  google {
    project     = "my-project"
    credentials = "my-credentials"
  }
}
