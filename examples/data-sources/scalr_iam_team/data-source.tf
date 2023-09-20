data "scalr_iam_team" "example1" {
  id         = "team-xxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_iam_team" "example2" {
  name       = "dev"
  account_id = "acc-xxxxxxx"
}
