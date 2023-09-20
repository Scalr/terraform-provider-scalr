# To retrieve a custom role, an account id and role id (or name) are required:

data "scalr_role" "example1" {
  id         = "role-xxxxxxx"
  account_id = "acc-xxxxxxx"
}

data "scalr_role" "example2" {
  name       = "WorkspaceAdmin"
  account_id = "acc-xxxxxxx"
}

# To retrieve system-managed roles an account id has to be omitted:

data "scalr_role" "example3" {
  name = "user"
}
