package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalrProviderConfigurationDataSource_name(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationAwsDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config: testAccScalrProviderConfigurationAwsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_provider_configuration.aws", "scalr_provider_configuration.aws"),
				),
			},
		},
	})
}
func TestAccScalrProviderConfigurationDataSource_provider_type(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationGoogleDataSourceInitConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_provider_configuration.google", "scalr_provider_configuration.google"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationGoogleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_provider_configuration.google", "scalr_provider_configuration.google"),
				),
			},
		},
	})
}

var testAccScalrProviderConfigurationAwsDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "google_pcfg"
  account_id = "%[1]s"
  google {
    project     = "my-new-project"
    credentials = "my-new-credentials"
  }
}
resource "scalr_provider_configuration" "aws" {
  name                   = "aws_pcfg"
  account_id             = "%[1]s"
  aws {
    secret_key = "my-new-secret-key"
    access_key = "my-new-access-key"
  }
}`, defaultAccount)
var testAccScalrProviderConfigurationAwsDataSourceConfig = testAccScalrProviderConfigurationAwsDataSourceInitConfig + `
data "scalr_provider_configuration" "aws" {
  name = scalr_provider_configuration.aws.name
}
`

var testAccScalrProviderConfigurationGoogleDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "google_pcfg"
  account_id = "%[1]s"
  google {
    project     = "my-new-project"
    credentials = "my-new-credentials"
  }
}
resource "scalr_provider_configuration" "aws" {
  name                   = "aws_pcfg"
  account_id             = "%[1]s"
  aws {
    secret_key = "my-new-secret-key"
    access_key = "my-new-access-key"
  }
}`, defaultAccount)
var testAccScalrProviderConfigurationGoogleDataSourceConfig = testAccScalrProviderConfigurationGoogleDataSourceInitConfig + `
data "scalr_provider_configuration" "google" {
	account_id    = "%[1]s"
	provider_type = "google"
}`
