package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccScalrProviderConfigurationsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
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

func TestAccScalrProviderConfigurationsDataSource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationsDataSourceTagsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_provider_configurations.tagged_foo_bar", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.scalr_provider_configurations.tagged_baz", "ids.#", "2"),
				),
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

func testAccScalrProviderConfigurationsDataSourceTagsConfig() string {
	return fmt.Sprintf(`
resource "scalr_tag" "foo" {
  name = "pcfg-foo"
}

resource "scalr_tag" "bar" {
  name = "pcfg-bar"
}

resource "scalr_tag" "baz" {
  name = "pcfg-baz"
}

resource "scalr_provider_configuration" "tagged_foobar" {
  name       = "tagged-foobar"
  account_id = "%[1]s"
  tag_ids    = [scalr_tag.foo.id, scalr_tag.bar.id]
  custom {
    provider_name = "kubernetes"
    argument {
      name  = "host"
      value = "my-host"
    }
  }
}

resource "scalr_provider_configuration" "tagged_barbaz" {
  name       = "tagged-barbaz"
  account_id = "%[1]s"
  tag_ids    = [scalr_tag.bar.id, scalr_tag.baz.id]
  custom {
    provider_name = "kubernetes"
    argument {
      name  = "host"
      value = "my-host2"
    }
  }
}

resource "scalr_provider_configuration" "tagged_baz" {
  name       = "tagged-baz"
  account_id = "%[1]s"
  tag_ids    = [scalr_tag.baz.id]
  custom {
    provider_name = "consul"
    argument {
      name  = "address"
      value = "demo.consul.io:80"
    }
  }
}

data "scalr_provider_configurations" "tagged_foo_bar" {
  tag_ids = [scalr_tag.foo.id, scalr_tag.bar.id]
  depends_on = [
    scalr_provider_configuration.tagged_foobar,
    scalr_provider_configuration.tagged_barbaz,
    scalr_provider_configuration.tagged_baz,
  ]
}

data "scalr_provider_configurations" "tagged_baz" {
  tag_ids = [scalr_tag.baz.id]
  depends_on = [
    scalr_provider_configuration.tagged_foobar,
    scalr_provider_configuration.tagged_barbaz,
    scalr_provider_configuration.tagged_baz,
  ]
}
`, defaultAccount)
}
