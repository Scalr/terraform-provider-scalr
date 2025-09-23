output "url" {
  description = "The URL of the API Gateway endpoint"
  value       = "https://${aws_api_gateway_rest_api.scalr_webhook.id}.execute-api.${data.aws_region.current.name}.amazonaws.com/${aws_api_gateway_stage.prod.stage_name}/trigger"
}

output "api_key" {
  description = "The API key for authentication"
  value       = aws_api_gateway_api_key.scalr_webhook_key.value
  sensitive   = true
}

output "headers" {
  description = "The headers required for API requests"
  value = {
    "x-api-key" = aws_api_gateway_api_key.scalr_webhook_key.value
  }
  sensitive = true
}
