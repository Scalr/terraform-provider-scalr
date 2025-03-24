package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var rName = acctest.RandomWithPrefix("test-pcfg")

func TestAccScalrProviderConfigurationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config:      `data scalr_provider_configuration test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_provider_configuration test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrProviderConfigurationDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_provider_configuration.kubernetes", "scalr_provider_configuration.kubernetes"),
					testAccCheckEqualID("data.scalr_provider_configuration.consul", "scalr_provider_configuration.consul"),
					testAccCheckEqualID("data.scalr_provider_configuration.consul_id", "scalr_provider_configuration.consul"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_provider_configuration.kubernetes", "name",
						"scalr_provider_configuration.kubernetes", "name",
					),
					resource.TestCheckResourceAttr(
						"data.scalr_provider_configuration.kubernetes", "provider_name", "kubernetes",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_provider_configuration.kubernetes", "owners",
						"scalr_provider_configuration.kubernetes", "owners",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_provider_configuration.kubernetes", "environments",
						"scalr_provider_configuration.kubernetes", "environments",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_provider_configuration.consul", "name",
						"scalr_provider_configuration.consul", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_provider_configuration.consul_id", "name",
						"scalr_provider_configuration.consul", "name",
					),
				),
			},
			{
				Config: testAccScalrProviderConfigurationDataSourceScalrConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_provider_configuration.scalr", "scalr_provider_configuration.scalr"),
					resource.TestCheckResourceAttr("data.scalr_provider_configuration.scalr", "name", rName),
					resource.TestCheckResourceAttr("data.scalr_provider_configuration.scalr", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("data.scalr_provider_configuration.scalr", "provider_name", "scalr"),
				),
			},
			{
				Config: testAccScalrProviderConfigurationDataSourceInitConfig,
			},
		},
	})
}

func TestAccScalrProviderConfigurationDataSource_UpgradeFromSDK(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"scalr": {
						Source:            "registry.scalr.io/scalr/scalr",
						VersionConstraint: "<=2.5.0",
					},
				},
				Config: testAccScalrProviderConfigurationDataSourceInitConfig,
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"scalr": {
						Source:            "registry.scalr.io/scalr/scalr",
						VersionConstraint: "<=2.5.0",
					},
				},
				Config: testAccScalrProviderConfigurationDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_provider_configuration.kubernetes", "id"),
					resource.TestCheckResourceAttrSet("data.scalr_provider_configuration.consul", "id"),
					resource.TestCheckResourceAttrSet("data.scalr_provider_configuration.consul_id", "id"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(t),
				Config:                   testAccScalrProviderConfigurationDataSourceConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

var testAccScalrProviderConfigurationDataSourceScalrConfig = testAccScalrProviderConfigurationScalrConfig(rName) + `
data "scalr_provider_configuration" "scalr" {
	  name = scalr_provider_configuration.scalr.name
}`

var testAccScalrProviderConfigurationDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name       = "%[1]s-kubernetes1"
  account_id = "%[2]s"
  owners      = [scalr_iam_team.test.id]
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

resource "scalr_iam_team" "test" {
	name        = "test-pcfg-data-source-owner"
	description = "Test team"
	users       = []
  }

resource "scalr_provider_configuration" "consul" {
  name       = "%[1]s-consul"
  account_id = "%[2]s"
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
}
`, rName, defaultAccount)

var testAccScalrProviderConfigurationDataSourceConfig = testAccScalrProviderConfigurationDataSourceInitConfig + `
data "scalr_provider_configuration" "kubernetes" {
  name = scalr_provider_configuration.kubernetes.name
}
data "scalr_provider_configuration" "consul" {
  provider_name = "consul"
}
data "scalr_provider_configuration" "consul_id" {
  id = scalr_provider_configuration.consul.id
}
`
