data "scalr_vcs_provider" "example" {
  name       = "vcs-name"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_environment" "example" {
  name       = "env-name"
  account_id = "acc-xxxxxxxxxx"
}

resource "scalr_workspace" "example" {
  name            = "my-workspace-name"
  environment_id  = data.scalr_environment.example.id
  vcs_provider_id = data.scalr_vcs_provider.example.id

  working_directory = "example/path"

  vcs_repo {
    identifier       = "org/repo"
    branch           = "dev"
    trigger_prefixes = ["stage", "prod"]
  }

  provider_configuration {
    id    = "pcfg-xxxxxxxxxx"
    alias = "us_east1"
  }
  provider_configuration {
    id    = "pcfg-yyyyyyyyyy"
    alias = "us_east2"
  }
}

resource "scalr_workspace" "trigger_patterns" {
  name            = "trigger_patterns"
  environment_id  = data.scalr_environment.example.id
  vcs_provider_id = data.scalr_vcs_provider.example.id

  working_directory = "example/path"

  vcs_repo {
    identifier       = "org/repo"
    branch           = "dev"
    trigger_patterns = <<-EOT
    !*.MD
    !/**/test/

    /infrastructure/environments/pre/
    /infrastructure/modules/app/
    /infrastructure/modules/db/
    /infrastructure/modules/queue/
    EOT
  }
}
