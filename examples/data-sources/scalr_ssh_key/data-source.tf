data "scalr_ssh_key" "example1" {
  id         = "ssh-xxxxxxxxxx"
  account_id = "acc-xxxxxxxxxx"
}

data "scalr_ssh_key" "example2" {
  name       = "ssh_key_name"
  account_id = "acc-xxxxxxxxxx"
}
