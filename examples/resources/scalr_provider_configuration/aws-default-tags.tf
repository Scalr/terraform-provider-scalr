resource "scalr_provider_configuration" "aws_tags" {
  name         = "aws_stage_us_east_1"
  account_id   = "acc-xxxxxxxxxx"
  environments = ["*"]
  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    secret_key       = "my-secret-key"
    access_key       = "my-access-key"
    default_tags {
      tags = {
        Environment = "Staging"
        Owner       = "QATeam"
      }
      strategy = "update"
    }
  }
}
