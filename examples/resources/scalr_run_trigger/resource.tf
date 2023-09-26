data "scalr_workspace" "downstream" {
  name           = "downstream"
  environment_id = "env-xxxxxxxxxx"
}

data "scalr_workspace" "upstream" {
  name           = "upstream"
  environment_id = "env-xxxxxxxxxx"
}

resource "scalr_run_trigger" "set_downstream" {
  # run automatically triggered in this workspace once the run in the upstream workspace is applied
  downstream_id = data.scalr_workspace.downstream.id
  upstream_id   = data.scalr_workspace.upstream.id
}
