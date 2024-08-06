resource "scalr_event_bridge_integration" "example" {
  name           = "via-provider-aws-bridge"
  aws_account_id = "111267354555"
  region         = "us-east-1"
}
