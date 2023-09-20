resource "scalr_role" "reader" {
  name        = "Reader"
  account_id  = "acc-xxxxxxxx"
  description = "Read access to all resources."

  permissions = [
    "*:read",
  ]
}

resource "scalr_access_policy" "team_read_all_on_acc_scope" {
  subject {
    type = "team"
    id   = "team-xxxxxxx"
  }
  scope {
    type = "account"
    id   = "acc-xxxxxxx"
  }

  role_ids = [
    scalr_role.reader.id
  ]
}
