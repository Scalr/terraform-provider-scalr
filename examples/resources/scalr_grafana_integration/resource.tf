# Grafana Cloud
resource "scalr_grafana_integration" "example" {
  name      = "grafana-cloud"
  kind      = "cloud"
  endpoint  = "https://otlp-gateway-prod-eu-west-2.grafana.net/otlp"
  auth_type = "basic"
  username  = "123456" # Grafana Cloud instance ID
  secret    = var.grafana_cloud_token
}

# Self-hosted multi-tenant Loki
resource "scalr_grafana_integration" "loki" {
  name      = "loki"
  kind      = "loki"
  endpoint  = "https://loki.example.com/otlp"
  auth_type = "bearer"
  secret    = var.loki_token
  tenant_id = "scalr"
}
