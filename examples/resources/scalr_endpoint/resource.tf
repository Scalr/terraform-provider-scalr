resource "scalr_endpoint" "example" {
  name           = "my-endpoint-name"
  secret_key     = "my-secret-key"
  timeout        = 15
  max_attempts   = 3
  url            = "https://my-endpoint.url"
  environment_id = "env-xxxxxxxxxxxx"
}
