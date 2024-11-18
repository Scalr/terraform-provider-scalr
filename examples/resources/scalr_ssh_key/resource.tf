resource "scalr_ssh_key" "example" {
  name        = "example-ssh-key"
  private_key = <<EOF
-----BEGIN OPENSSH PRIVATE KEY-----
YOUR_PRIVATE_KEY_CONTENT_HERE
-----END OPENSSH PRIVATE KEY-----
EOF

  account_id   = "acc-xxxxxxxxxx"
  environments = ["env-xxxxxxxxxx"]
}