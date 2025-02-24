resource "scalr_integration_infracost" "example" {
  name         = "infracost"
  api_key      = "ico-xxxxx"
  environments = ["*"]
}