resource "scalr_role" "writer" {
  name        = "Writer"
  account_id  = "acc-xxxxxxxx"
  description = "Write access to all resources."

  permissions = [
    "*:update",
    "*:delete",
    "*:create"
  ]
}
