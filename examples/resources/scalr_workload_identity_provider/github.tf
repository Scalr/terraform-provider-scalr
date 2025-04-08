resource "scalr_workload_identity_provider" "github" {
  name              = "github-actions"
  url               = "https://token.actions.githubusercontent.com"
  allowed_audiences = ["scalr-github-actions"]
}
