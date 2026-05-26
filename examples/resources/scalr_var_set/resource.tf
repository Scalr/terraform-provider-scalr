resource "scalr_var_set" "example" {
  name         = "my-var-set-1"
  description  = "Var set description"
  environments = ["env-xxxxxxxxxx", "env-yyyyyyyyyy"]
}

resource "scalr_var_set" "example_shared" {
  name         = "my-var-set-2"
  description  = "Var set description"
  environments = ["*"]
  owners       = ["team-xxxxxxxxxx", "team-yyyyyyyyyy"]
}
