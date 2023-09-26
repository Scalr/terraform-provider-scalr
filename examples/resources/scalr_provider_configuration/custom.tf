resource "scalr_provider_configuration" "kubernetes" {
  name       = "k8s"
  account_id = "acc-xxxxxxxxxx"
  custom {
    provider_name = "kubernetes"
    argument {
      name        = "host"
      value       = "my-host"
      description = "The hostname (in form of URI) of the Kubernetes API."
    }
    argument {
      name  = "username"
      value = "my-username"
    }
    argument {
      name      = "password"
      value     = "my-password"
      sensitive = true
    }
  }
}
