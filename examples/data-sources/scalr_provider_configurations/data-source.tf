data "scalr_provider_configurations" "aws" {
  name = "in:aws_dev,aws_demo,aws_prod"
}

data "scalr_provider_configurations" "google" {
  provider_name = "google"
}
