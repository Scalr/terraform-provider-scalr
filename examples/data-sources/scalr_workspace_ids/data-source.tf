data "scalr_workspace_ids" "app-frontend" {
  names          = ["app-frontend-prod", "app-frontend-dev1", "app-frontend-staging"]
  environment_id = "env-xxxxxxxxxx"
}

data "scalr_workspace_ids" "all" {
  names          = ["*"]
  environment_id = "env-xxxxxxxxxx"
}
