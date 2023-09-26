resource "scalr_role" "writer" {
  name        = "Writer"
  account_id  = "acc-xxxxxxxxxx"
  description = "Write access to all resources."

  permissions = [
    "*:update",
    "*:delete",
    "*:create"
  ]
}
