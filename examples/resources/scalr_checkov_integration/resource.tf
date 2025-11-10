resource "scalr_checkov_integration" "example1" {
  name         = "my-checkov-integration-1"
  environments = ["*"]
  cli_args     = "--quiet"
}

resource "scalr_checkov_integration" "example2" {
  name = "my-checkov-integration-2"
  vcs_repo {
    identifier = "org/repo"
    branch     = "main"
  }
}

resource "scalr_checkov_integration" "example3" {
  name         = "my-checkov-integration-3"
  environments = []
  cli_args     = "--compact"
}
