package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalrProviderConfigurationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrProviderConfigurationDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config:      `data scalr_provider_configuration test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_provider_configuration test {name = ""}`,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrProviderConfigurationDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_provider_configuration.kubernetes", "scalr_provider_configuration.kubernetes"),
					testAccCheckEqualID("data.scalr_provider_configuration.consul", "scalr_provider_configuration.consul"),
					testAccCheckEqualID("data.scalr_provider_configuration.consul_id", "scalr_provider_configuration.consul"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_provider_configuration.kubernetes", "owners",
						"scalr_provider_configuration.kubernetes", "owners",
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

var rName = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

var testAccScalrProviderConfigurationDataSourceScalrConfig = testAccScalrProviderConfigurationScalrConfig(rName) + `
data "scalr_provider_configuration" "scalr" {
	  name = scalr_provider_configuration.scalr.name
}`

var testAccScalrProviderConfigurationDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name       = "kubernetes1"
  account_id = "%[1]s"
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
}
`, defaultAccount)

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
