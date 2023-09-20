resource "scalr_iam_team" "example" {
  name        = "dev"
  description = "Developers"
  account_id  = "acc-xxxxxxxx"

  users = ["user-xxxxxxxx", "user-yyyyyyyy"]
}
