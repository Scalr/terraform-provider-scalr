output "agent_pool_id" {
  description = "The ID of the created Scalr agent pool"
  value       = module.agent_pool.agent_pool_id
}

output "agent_token" {
  description = "The token for the Scalr agent"
  value       = module.agent_pool.agent_token
  sensitive   = true
}

output "webhook_url" {
  description = "The URL of the API Gateway endpoint"
  value       = module.api_gateway.url
}

output "api_key" {
  description = "The API key for authentication"
  value       = module.api_gateway.api_key
  sensitive   = true
}
