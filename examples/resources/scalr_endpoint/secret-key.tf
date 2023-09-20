resource "random_string" "r" {
  length = 16
}

resource "scalr_endpoint" "example" {
  # ...
  secret_key = random_string.r.result
  # ...
}
