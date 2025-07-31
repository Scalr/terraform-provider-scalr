resource "scalr_provider_configuration" "elasticstack" {
  name       = "elastic"
  account_id = "acc-xxxxxxxxxx"
  custom {
    provider_name = "elasticstack"
    argument {
      name        = "endpoints"
      value       = "[\"https://elasticsearch.example.com:9200\", \"https://elasticsearch2.example.com:9200\"]"
      description = "List of Elasticsearch endpoints."
      hcl         = true
    }
    argument {
      name        = "username"
      value       = "elastic"
      description = "Username for Elasticsearch authentication."
    }
    argument {
      name        = "password"
      value       = "my-elastic-password"
      sensitive   = true
      description = "Password for Elasticsearch authentication."
    }
  }
} 