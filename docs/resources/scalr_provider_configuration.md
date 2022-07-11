---
layout: "scalr"
page_title: "Scalr: scalr_provider_configuration"
sidebar_current: "docs-resource-scalr-provider-configuration"
description: |-
  Manages provider configurations.
---

# scalr_provider_configuration Resource

Manage the state of provider configuraitons in Scalr. Creates, updates and destroy.

## Example Usage

Aws provider:

```hcl
resource "scalr_provider_configuration" "aws" {
  name                   = "aws_dev_us_east_1"
  account_id             = "acc-xxxxxxxxx"
  export_shell_variables = false
  environments           = [scalr_environment.env1.id]
  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    secret_key       = "my-secret-key"
    access_key       = "my-access-key"
  }
}
```

Google provider:

```hcl
resource "scalr_provider_configuration" "google" {
  name       = "google_main"
  account_id = "acc-xxxxxxxxx"
  google {
    project     = "my-project"
    credentials = "my-credentials"
  }
}
```

Azurerm provider:

```hcl
resource "scalr_provider_configuration" "azurerm" {
  name       = "azurerm"
  account_id = "acc-xxxxxxxxx"
  azurerm {
    client_id       = "my-client-id"
    client_secret   = "my-client-secret"
    subscription_id = "my-subscription-id"
    tenant_id       = "my-tenant-id"
  }
}
```

Scalr provider:

```hcl
resource "scalr_provider_configuration" "scalr" {
  name       = "scalr"
  account_id = "acc-xxxxxxxxx"
  scalr {
    hostname       = "scalr.host.example.com"
    token          = "my-scalr-token"
  }
}
```

Other providers:

```hcl
resource "scalr_provider_configuration" "kubernetes" {
  name                   = "k8s"
  account_id             = "acc-xxxxxxxxx"
  custom {
    provider_name = "kubernetes"
    argument {
      name        = "host"
      value       = "my-host"
      description = "The hostname (in form of URI) of the Kubernetes API."
    }
    argument {
      name  = "username"
      value = "my-username"
    }
    argument {
      name      = "password"
      value     = "my-password"
      sensitive = true
    }
  }
}
```

## Argument Reference

* `account_id` - (Required) The account that owns the variable, specified as an ID, in the format.
* `name` - (Required) The name of a Scalr provider configuration. This field is unique for the account.
* `export_shell_variables` - (Optional) Export provider variables into the run environment. This option is available only for built in providers.
* `environments` - (Optional) The list of environments attached to the provider configuration. Use `["*"]` to select all environments.
* `aws` - (Optional) Settings for the aws provider configuraiton. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `scalr`, `custom`.
   The `aws` block supports the following:
  * `account_type` - (Required) The type of AWS accoutn, available options: `regular`, `gov-cloud`, `cn-cloud`.
  * `credentials_type` - (Required) The type of AWS credentials, available options: `access_keys`, `role_delegation`.
  * `trusted_entity_type` - (Optional) Trusted entity type, available options: `aws_account`, `aws_service`. This option is required with `role_delegation` credentials type.
  * `role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role to assume. This option is required with `role_delegation` credentials type.
  * `external_id` - (Optional) External identifier to use when assuming the role. This option is required with `role_delegation` credentials type and `aws_account` trusted entity type.
  * `secret_key` - (Optional) AWS secret key. This option is required with `access_keys` credentials type.
  * `access_key` - (Optional) AWS access key.This option is required with `access_keys` credentials type.
* `google` - (Optional) Settings for the google provider configuraiton. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `scalr`, `custom`.
   The `google` block supports the following:
  * `project` - (Optional) The default project to manage resources in. If another project is specified on a resource, it will take precedence.
  * `credentials` - (Required) Service account key file in JSON format.
* `azurerm` - (Optional) Settings for the azurerm provider configuraiton. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `scalr`, `custom`.
   The `azurerm` block supports the following:
  * `client_id` - (Required) The Client ID which should be used.
  * `client_secret` - (Required) The Client Secret which should be used.
  * `tenant_id` - (Required) The Tenant ID should be used.
  * `subscription_id` - (Optional) The Subscription ID which should be used.
* `scalr` - (Optional) Settings for the Scalr provider configuraiton. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `scalr`, `custom`.
  The `scalr` block supports the following:
    * `hostname` - (Optional) The Scalr hostname which should be used.
    * `token` - (Optional) The Scalr Token which should be used.
* `custom` - (Optional) Settings for the provider configuraiton that does not have first class scalr support. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `scalr`, `custom`.
   The `custom` block supports the following:
  * `provider_name` - (Required) The name of a Terraform provider.
  * `argument` - (Required) The provider configuration argument.
     The `argument` block supports the following:
    * `name` - (Required) The name of the provider configuration argument. 
    * `value` - (Optional) The value of the provider configuration argument.
    * `sensitive` - (Optional) Set (true/false) to configure as sensitive. Default `false`.
    * `description` - (Optional) The description of the provider configuration argument.


## Attribute Reference

All arguments plus:

* `id` - The ID of the provider configuration, in the format `pcfg-xxxxxxxx`.
