data "scalr_module_versions" "example1" {
  id = "mod-xxxxxxxxxx"
}

data "scalr_module_versions" "example2" {
  source = "env-xxxxxxxxxx/module-name/scalr"
}

data "scalr_module_versions" "example3" {
  id     = "mod-xxxxxxxxxx"
  source = "env-xxxxxxxxxx/module-name/scalr"
}
