package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalrAgentPoolDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolAccountDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_agent_pool.test", "scalr_agent_pool.test"),
					resource.TestCheckResourceAttrSet("data.scalr_agent_pool.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "name", "agent_pool-test"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},
		},
	})
}
func TestAccScalrAgentPoolDataSource_basic_env(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolEnvDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_agent_pool.test", "scalr_agent_pool.test"),
					resource.TestCheckResourceAttrSet("data.scalr_agent_pool.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "name", "agent_pool-test"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func testAccScalrAgentPoolAccountDataSourceConfig() string {
	return fmt.Sprintf(`
resource "scalr_agent_pool" "test" {
  name             = "agent_pool-test"
  account_id       = "%s"
}

data "scalr_agent_pool" "test" {
	name = scalr_agent_pool.test.name
	account_id = scalr_agent_pool.test.account_id
}`, defaultAccount)
}

func testAccScalrAgentPoolEnvDataSourceConfig() string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name             = "agent_pool-test-123"
  account_id       = "%s"

}
resource "scalr_agent_pool" "test" {
  name             = "agent_pool-test"
  account_id       = "%s"
  environment_id = scalr_environment.test.id
}

data "scalr_agent_pool" "test" {
	name = scalr_agent_pool.test.name
	account_id = scalr_agent_pool.test.account_id
	environment_id = scalr_environment.test.id
}`, defaultAccount, defaultAccount)
}
