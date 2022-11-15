package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccScalrProviderConfigurationsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationsDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config: testAccScalrProviderConfigurationsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProviderConfigurationsDataSourceNameFilter(),
					testAccCheckProviderConfigurationsDataSourceTypeFilter(),
				),
			},
			{
				Config: testAccScalrProviderConfigurationsDataSourceInitConfig, // depends_on works improperly with data sources
			},
		},
	})
}

func testAccCheckProviderConfigurationsDataSourceNameFilter() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var expectedIds []string
		resourceNames := []string{"kubernetes2", "consul"}
		for _, name := range resourceNames {
			rsName := "scalr_provider_configuration." + name
			rs, ok := s.RootModule().Resources[rsName]
			if !ok {
				return fmt.Errorf("Not found: %s", rsName)
			}
			expectedIds = append(expectedIds, rs.Primary.ID)

		}
		dataSource, ok := s.RootModule().Resources["data.scalr_provider_configurations.kubernetes2consul"]
		if !ok {
			return fmt.Errorf("Not found: data.scalr_provider_configurations.kubernetes2consul")
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
		resourceNames := []string{"kubernetes1", "kubernetes2"}
		for _, name := range resourceNames {
			rsName := "scalr_provider_configuration." + name
			rs, ok := s.RootModule().Resources[rsName]
			if !ok {
				return fmt.Errorf("Not found: %s", rsName)
			}
			expectedIds = append(expectedIds, rs.Primary.ID)

		}
		dataSource, ok := s.RootModule().Resources["data.scalr_provider_configurations.kubernetes"]
		if !ok {
			return fmt.Errorf("Not found: data.scalr_provider_configurations.kubernetes")
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

var testAccScalrProviderConfigurationsDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes1" {
  name       = "kubernetes1"
  account_id = "%[1]s"
  custom {
    provider_name = "kubernetes"
    argument {
      name  = "host"
      value = "my-host"
    }
    argument {
      name  = "username"
      value = "my-username"
    }
  }
}
resource "scalr_provider_configuration" "kubernetes2" {
  name       = "kubernetes2"
  account_id = "%[1]s"
  custom {
    provider_name = "kubernetes"
    argument {
      name  = "host"
      value = "my-host2"
    }
    argument {
      name  = "username"
      value = "my-username2"
    }
  }
}
resource "scalr_provider_configuration" "consul" {
  name       = "consul"
  account_id = "%[1]s"
  custom {
    provider_name = "consul"
    argument {
      name  = "address"
      value = "demo.consul.io:80"
    }
    argument {
      name  = "datacenter"
      value = "nyc1"
    }
  }
}`, defaultAccount)

var testAccScalrProviderConfigurationsDataSourceConfig = testAccScalrProviderConfigurationsDataSourceInitConfig + `
data "scalr_provider_configurations" "kubernetes2consul" {
  name = "in:kubernetes2,consul"
}
data "scalr_provider_configurations" "kubernetes" {
  provider_name = "kubernetes"
}
`
