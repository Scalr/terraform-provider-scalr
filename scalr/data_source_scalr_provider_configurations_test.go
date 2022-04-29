package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccScalrProviderConfigurationsDataSource_name(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationsAwsDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config: testAccScalrProviderConfigurationsAwsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProviderConfigurationsDataSourceNameFilter(),
				),
			},
		},
	})
}
func TestAccScalrProviderConfigurationsDataSource_provider_type(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationsGoogleDataSourceInitConfig,
			},
			{
				Config: testAccScalrProviderConfigurationsGoogleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProviderConfigurationsDataSourceTypeFilter(),
				),
			},
		},
	})
}
func testAccCheckProviderConfigurationsDataSourceNameFilter() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var expectedIds []string
		resourceNames := []string{"aws", "aws2"}
		for _, name := range resourceNames {
			rsName := "scalr_provider_configuration." + name
			rs, ok := s.RootModule().Resources[rsName]
			if !ok {
				return fmt.Errorf("Not found: %s", rsName)
			}
			expectedIds = append(expectedIds, rs.Primary.ID)

		}
		dataSource, ok := s.RootModule().Resources["data.scalr_provider_configurations.aws"]
		if !ok {
			return fmt.Errorf("Not found: data.scalr_provider_configurations.aws")
		}
		if dataSource.Primary.Attributes["ids.#"] != "2" {
			return fmt.Errorf("Bad provider configuration ids, expected: %#v, got: %#v", expectedIds, dataSource.Primary.Attributes["ids"])
		}

		resultIds := []string{dataSource.Primary.Attributes["ids.0"], dataSource.Primary.Attributes["ids.1"]}

		for _, expectedId := range expectedIds {
			found := false
			for _, resultId := range resultIds {
				if resultId == expectedId {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("Bad provider configuration ids, expected: %#v, got: %#v", expectedIds, resultIds)
			}
		}

		return nil
	}
}

func testAccCheckProviderConfigurationsDataSourceTypeFilter() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var expectedIds []string
		resourceNames := []string{"google", "google2"}
		for _, name := range resourceNames {
			rsName := "scalr_provider_configuration." + name
			rs, ok := s.RootModule().Resources[rsName]
			if !ok {
				return fmt.Errorf("Not found: %s", rsName)
			}
			expectedIds = append(expectedIds, rs.Primary.ID)

		}
		dataSource, ok := s.RootModule().Resources["data.scalr_provider_configurations.google"]
		if !ok {
			return fmt.Errorf("Not found: data.scalr_provider_configurations.google")
		}
		if dataSource.Primary.Attributes["ids.#"] != "2" {
			return fmt.Errorf("Bad provider configuration ids, expected: %#v, got: %#v", expectedIds, dataSource.Primary.Attributes["ids"])
		}

		resultIds := []string{dataSource.Primary.Attributes["ids.0"], dataSource.Primary.Attributes["ids.1"]}

		for _, expectedId := range expectedIds {
			found := false
			for _, resultId := range resultIds {
				if resultId == expectedId {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("Bad provider configuration ids, expected: %#v, got: %#v", expectedIds, resultIds)
			}
		}

		return nil
	}
}

var testAccScalrProviderConfigurationsAwsDataSourceInitConfig = fmt.Sprintf(`
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
}
resource "scalr_provider_configuration" "aws2" {
  name                   = "aws2_pcfg"
  account_id             = "%[1]s"
  aws {
    secret_key = "my-new-secret-key"
    access_key = "my-new-access-key"
  }
}`, defaultAccount)
var testAccScalrProviderConfigurationsAwsDataSourceConfig = testAccScalrProviderConfigurationsAwsDataSourceInitConfig + `
data "scalr_provider_configurations" "aws" {
  name = "in:aws_pcfg,aws2_pcfg"
}
`

var testAccScalrProviderConfigurationsGoogleDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_provider_configuration" "google" {
  name       = "google_pcfg"
  account_id = "%[1]s"
  google {
    project     = "my-new-project"
    credentials = "my-new-credentials"
  }
}
resource "scalr_provider_configuration" "google2" {
  name       = "google2_pcfg"
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

var testAccScalrProviderConfigurationsGoogleDataSourceConfig = testAccScalrProviderConfigurationsGoogleDataSourceInitConfig + `
data "scalr_provider_configurations" "google" {
	provider_type = "google"
}`
