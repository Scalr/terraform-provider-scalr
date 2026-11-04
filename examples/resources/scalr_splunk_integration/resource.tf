resource "scalr_splunk_integration" "example" {
  name     = "splunk"
  endpoint = "https://http-inputs-acme.splunkcloud.com"
  token    = var.splunk_hec_token
  index    = "scalr_audit"
}

# Self-managed Splunk Enterprise, events go to the default index of the HEC token
resource "scalr_splunk_integration" "enterprise" {
  name     = "splunk-enterprise"
  endpoint = "https://splunk.example.com:8088"
  token    = var.splunk_enterprise_hec_token
}
