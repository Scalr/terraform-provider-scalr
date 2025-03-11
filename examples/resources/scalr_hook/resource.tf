resource "scalr_hook" "example" {
  name            = "hook-test"
  description     = "Hook description"
  interpreter     = "bash"
  scriptfile_path = "root.sh"
  vcs_provider_id = "vcs-xxxxx"
  account_id      = "acc-xxxxx"
  vcs_repo {
    identifier = "TestRepo/example"
    branch     = "main"
  }
}