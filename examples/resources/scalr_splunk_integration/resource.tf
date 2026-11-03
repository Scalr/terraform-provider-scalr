resource "scalr_splunk_integration" "example" {
  name     = "splunk"
  endpoint = "https://http-inputs-acme.splunkcloud.com"
  token    = var.splunk_hec_token
  index    = "scalr_audit"
}

# Self-managed Splunk Enterprise with a custom sourcetype
resource "scalr_splunk_integration" "enterprise" {
  name       = "splunk-enterprise"
  endpoint   = "https://splunk.example.com:8088"
  token      = var.splunk_enterprise_hec_token
  index      = "scalr_audit"
  source     = "scalr-prod"
  sourcetype = "scalr:audit"
}
