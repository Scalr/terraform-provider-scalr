resource "scalr_environment" "example" {
  name       = "my-environment"
  account_id = "acc-xxxxxxxxxx"
}

# Give the new environment access to an existing module namespace.
resource "scalr_module_namespace_environment" "example" {
  module_namespace_id = "modns-xxxxxxxxxx"
  environment_id      = scalr_environment.example.id
}
