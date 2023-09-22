data "scalr_workspace" "downstream" {
  name           = "downstream"
  environment_id = "env-xxxxxxxxx"
}

data "scalr_workspace" "upstream" {
  name           = "upstream"
  environment_id = "env-xxxxxxxxx"
}

resource "scalr_run_trigger" "set_downstream" {
  # run automatically triggered in this workspace once the run in the upstream workspace is applied
  downstream_id = data.scalr_workspace.downstream.id
  upstream_id   = data.scalr_workspace.upstream.id
}
