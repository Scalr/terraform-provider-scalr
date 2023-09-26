resource "scalr_iam_team" "example" {
  name        = "dev"
  description = "Developers"
  account_id  = "acc-xxxxxxxxxx"

  users = ["user-xxxxxxxxxx", "user-yyyyyyyyyy"]
}
