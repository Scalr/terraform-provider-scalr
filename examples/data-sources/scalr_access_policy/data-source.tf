data "scalr_access_policy" "example" {
  id = "ap-xxxxxxxxxx"
}

output "scope_id" {
  value = data.scalr_access_policy.example.scope[0].id
}

output "subject_id" {
  value = data.scalr_access_policy.example.subject[0].id
}
