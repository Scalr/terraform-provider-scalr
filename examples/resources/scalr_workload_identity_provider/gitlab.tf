resource "scalr_workload_identity_provider" "gitlab" {
  name              = "gitlab-ci"
  url               = "https://gitlab.com"
  allowed_audiences = ["scalr-gitlab-ci"]
}
