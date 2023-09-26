data "scalr_environment" "example" {
  name       = "env-name"
  account_id = "acc-xxxxxxxxxx"
}

locals {
  modules = {
    "${data.scalr_environment.example.id}" : "module-name/provider",         # environment-level module will be selected
    "${data.scalr_environment.example.account_id}" : "module-name/provider", # account-level module will be selected
  }
}

data "scalr_module_version" "example" {
  for_each = local.modules
  source   = "${each.key}/${each.value}"
}

resource "scalr_workspace" "example" {
  for_each       = data.scalr_module_version.example
  environment_id = data.scalr_environment.example.id

  name              = replace(each.value.source, "/", "-")
  module_version_id = each.value.id
}
