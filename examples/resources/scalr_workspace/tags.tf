resource "scalr_tag" "team-a" {
  name = "TeamA"
}

resource "scalr_tag" "team-b" {
  name = "TeamB"
}

resource "scalr_workspace" "example-a" {
  environment_id = "env-xxxxxxxxxx"
  name           = "example-a"
  tag_ids        = [scalr_tag.team-a.id]
}

resource "scalr_workspace" "example-b" {
  environment_id = "env-xxxxxxxxxx"
  name           = "example-b"
  tag_ids        = [scalr_tag.team-b.id]
}
