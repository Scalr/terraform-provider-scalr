data "scalr_environments" "exact-names" {
  name = "in:production,development"
}

data "scalr_environments" "app-frontend" {
  name = "like:app-frontend-"
}

data "scalr_environments" "tagged" {
  tag_ids = ["tag-xxxxxxxxxx", "tag-yyyyyyyyyy"]
}

data "scalr_environments" "all" {
  account_id = "acc-xxxxxxxxxx"
}
