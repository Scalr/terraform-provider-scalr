data "scalr_iam_user" "example1" {
  id = "user-xxxxxxxxxx"
}

data "scalr_iam_user" "example2" {
  email = "user@test.com"
}
