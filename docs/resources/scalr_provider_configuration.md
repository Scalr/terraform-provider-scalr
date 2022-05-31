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
  aws {
    secret_key = "my-secret-key"
    access_key = "my-access-key"
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
* `aws` - (Optional) Settings for the aws provider configuraiton. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `custom`.
   The `aws` block supports the following:
  * `secret_key` - (Optional) AWS secret key. 
  * `access_key` - (Optional) AWS access key.
* `google` - (Optional) Settings for the google provider configuraiton. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `custom`.
   The `google` block supports the following:
  * `project` - (Optional) The default project to manage resources in. If another project is specified on a resource, it will take precedence.
  * `credentials` - (Optional) Either the path to or the contents of a service account key file in JSON format. You can manage key files using the Cloud Console. If not provided, the application default credentials will be used.
* `azurerm` - (Optional) Settings for the azurerm provider configuraiton. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `custom`.
   The `azurerm` block supports the following:
  * `client_id` - (Optional) The Client ID which should be used.
  * `client_secret` - (Optional) The Client Secret which should be used.
  * `subscription_id` - (Optional) The Subscription ID which should be used. 
  * `tenant_id` - (Optional) The Tenant ID should be used.
* `scalr` - (Optional) Settings for the Scalr provider configuraiton. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `scalr`, `custom`.
  The `scalr` block supports the following:
    * `hostname` - (Optional) The Scalr hostname which should be used.
    * `token` - (Optional) The Scalr Token which should be used.
* `custom` - (Optional) Settings for the provider configuraiton that does not have first class scalr support. Exactly one of the following attributes must be set: `aws`, `google`, `azurerm`, `custom`.
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
