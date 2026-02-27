# Use two AWS provider configurations with the same alias: one for plan phase,
# another for apply phase. When apply_only is enabled, the provider configuration
# is used only during the apply phase of the run.

# AWS provider configuration used during plan phase (default)
resource "scalr_provider_configuration" "aws_plan" {
  name                   = "aws_plan_us_east_1"
  account_id             = "acc-xxxxxxxxxx"
  export_shell_variables = false
  environments           = ["env-xxxxxxxxxx"]

  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    access_key       = "my-plan-access-key"
    secret_key       = "my-plan-secret-key"
  }
}

# AWS provider configuration used only during apply phase
resource "scalr_provider_configuration" "aws_apply" {
  name                   = "aws_apply_us_east_1"
  account_id             = "acc-xxxxxxxxxx"
  export_shell_variables = false
  environments           = ["env-xxxxxxxxxx"]
  apply_only             = true

  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    access_key       = "my-apply-access-key"
    secret_key       = "my-apply-secret-key"
  }
}

# Workspace with both provider configurations linked under the same alias
resource "scalr_workspace" "example" {
  name            = "plan-apply-aws-example"
  environment_id  = "env-xxxxxxxxxx"
  vcs_provider_id = "vcs-xxxxxxxxxx"

  vcs_repo {
    identifier = "org/repo"
    branch     = "main"
  }

  provider_configuration {
    id    = scalr_provider_configuration.aws_plan.id
    alias = "us_east_1"
  }
  provider_configuration {
    id    = scalr_provider_configuration.aws_apply.id
    alias = "us_east_1"
  }
}
