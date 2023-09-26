data "scalr_iam_team" "example1" {
  id         = "team-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_iam_team" "example2" {
  name       = "dev"
  account_id = "acc-xxxxxxxxxx"
}
