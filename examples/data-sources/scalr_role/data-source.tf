# To retrieve a custom role, an account id and role id (or name) are required:

data "scalr_role" "example1" {
  id         = "role-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_role" "example2" {
  name       = "WorkspaceAdmin"
  account_id = "acc-xxxxxxxxxx"
}

# To retrieve system-managed roles an account id has to be omitted:

data "scalr_role" "example3" {
  name = "user"
}
