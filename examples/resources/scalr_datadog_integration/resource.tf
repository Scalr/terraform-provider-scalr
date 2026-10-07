resource "scalr_datadog_integration" "example" {
  name           = "datadog"
  api_key        = var.datadog_api_key
  deployment_url = "https://us5.datadoghq.com"
}
