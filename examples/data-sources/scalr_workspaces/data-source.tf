data "scalr_workspaces" "exact-names" {
  name = "in:production,development"
}

data "scalr_workspaces" "app-frontend" {
  name           = "like:app-frontend-"
  environment_id = "env-xxxxxxxxxxx"
}

data "scalr_workspaces" "tagged" {
  tag_ids = ["tag-xxxxxxxxxxx", "tag-yyyyyyyyyyy"]
}

data "scalr_workspaces" "all" {
  environment_id = "env-xxxxxxxxxxx"
}
