
# Data Source `scalr_iam_team` 

Retrieves the details of a Scalr team.

## Example Usage

```hcl
data "scalr_iam_team" "example" {
  id         = "team-xxxxxxx"
  account_id = "acc-xxxxxxx"
}
```

```hcl
data "scalr_iam_team" "example" {
  name       = "dev"
  account_id = "acc-xxxxxxx"
}
```

## Argument Reference

* `id` - (Optional) Identifier of the team.
* `name` - (Optional) Name of the team.
* `account_id` - (Optional) The identifier of the Scalr account.

Arguments `id` and `name` are both optional, specify at least one of them to obtain `scalr_iam_team`.

## Attribute Reference

All arguments plus:

* `description` - A verbose description of the team.
* `identity_provider_id` - An identifier of an identity provider team is linked to, in the format `idp-<RANDOM STRING>`.
* `users` - The list of the user identifiers that belong to the team.
