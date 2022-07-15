---
layout: "scalr"
page_title: "Scalr: scalr_provider_configuration"
sidebar_current: "docs-resource-scalr-provider-configuration"
description: |-
  Manages provider configurations.
---

# scalr_provider_configuration Resource

Provider configuration helps organizations to manage providers' secrets in a centralized way from a single place.
It natively supports the management of the major providers like Scalr, AWS, AzureRM, and Google Cloud Platform. 
But also, allows registering any custom provider. Please have a look at basic usage examples for each provider type.

## Basic Usage

### Scalr provider:

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

### Aws provider:

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

To get into more advanced AWS usage please refer to the official [AWS module](https://github.com/emocharnik/terraform-scalr-provider-configuration-aws).

## AzureRM provider:

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

### Google provider:

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

### Custom providers:

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
* `export_shell_variables` - (Optional) Export provider variables into the run environment. This option is available for built-in (Scalr, AWS, AzureRM, Google) providers only.
* `environments` - (Optional) The list of environment identifiers to the provider configuration is shared. Use `["*"]` to share with all environments.
* `scalr` - (Optional) Settings for the Scalr provider configuration. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.
  The `scalr` block supports the following:
    * `hostname` - (Optional) The Scalr hostname which should be used.
    * `token` - (Optional) The Scalr Token which should be used.
* `aws` - (Optional) Settings for the aws provider configuration. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.
   The `aws` block supports the following:
  * `account_type` - (Required) The type of AWS account, available options: `regular`, `gov-cloud`, `cn-cloud`.
  * `credentials_type` - (Required) The type of AWS credentials, available options: `access_keys`, `role_delegation`.
  * `trusted_entity_type` - (Optional) Trusted entity type, available options: `aws_account`, `aws_service`. This option is required with `role_delegation` credentials type.
  * `role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role to assume. This option is required with the `role_delegation` credentials type.
  * `external_id` - (Optional) External identifier to use when assuming the role. This option is required with `role_delegation` credentials type and `aws_account` trusted entity type.
  * `secret_key` - (Optional) AWS secret key. This option is required with `access_keys` credentials type.
  * `access_key` - (Optional) AWS access key. This option is required with `access_keys` credentials type.
* `google` - (Optional) Settings for the google provider configuration. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.
   The `google` block supports the following:
  * `credentials` - (Required) Service account key file in JSON format.
  * `project` - (Optional) The default project to manage resources in. If another project is specified on a resource, it will take precedence.
* `azurerm` - (Optional) Settings for the azurerm provider configuration. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.
   The `azurerm` block supports the following:
  * `client_id` - (Required) The Client ID which should be used.
  * `client_secret` - (Required) The Client Secret which should be used.
  * `tenant_id` - (Required) The Tenant ID should be used.
  * `subscription_id` - (Optional) The Subscription ID which should be used. If skipped, has to be set as a shell variable in the workspace or as a part of the source configuration.
* `custom` - (Optional) Settings for the provider configuration that does not have scalr support as a built-in provider. Exactly one of the following attributes must be set: `scalr`, `aws`, `google`, `azurerm`, `custom`.
   The `custom` block supports the following:
  * `provider_name` - (Required) The name of a Terraform provider.
  * `argument` - (Required) The provider configuration argument. Multiple instances are allowed per block.
     The `argument` block supports the following:
    * `name` - (Required) The name of the provider configuration argument. 
    * `value` - (Optional) The value of the provider configuration argument.
    * `sensitive` - (Optional) Set (true/false) to configure as sensitive. Default `false`.
    * `description` - (Optional) The description of the provider configuration argument.


## Attribute Reference

All arguments plus:

* `id` - The ID of the provider configuration, in the format `pcfg-xxxxxxxx`.
