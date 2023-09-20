resource "scalr_webhook" "example1" {
  name         = "my-webhook-1"
  enabled      = true
  url          = "https://my-endpoint.url"
  secret_key   = "my-secret-key"
  timeout      = 15
  max_attempts = 3
  events       = ["run:completed", "run:errored"]
  environments = ["env-xxxxxxxxxx"]
  header {
    name      = "header1"
    value     = "value1"
  }
  header {
    name      = "header2"
    value     = "value2"
  }
}

# Old-style webhook resource (deprecated):
resource "scalr_webhook" "example2" {
  name           = "my-webhook-2"
  enabled        = true
  endpoint_id    = "ep-xxxxxxxxxx"
  events         = ["run:completed", "run:errored"]
  workspace_id   = "ws-xxxxxxxxxx"
  environment_id = "env-xxxxxxxxxx"
}
