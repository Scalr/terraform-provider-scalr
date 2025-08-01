package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrAgentPoolDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_agent_pool test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_agent_pool test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_agent_pool test {name = ""}`,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrAgentPoolAccountDataSourceByIDConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_agent_pool.test", "scalr_agent_pool.test"),
					resource.TestCheckResourceAttrSet("data.scalr_agent_pool.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "name", "ds-agent_pool-test-acc"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr(
						"data.scalr_agent_pool.test", "environments.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_agent_pool.test",
						"environments.0",
						"scalr_environment.test-new",
						"id"),
				),
			},
			{
				Config: testAccScalrAgentPoolAccountDataSourceByNameConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_agent_pool.test", "scalr_agent_pool.test"),
					resource.TestCheckResourceAttrSet("data.scalr_agent_pool.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "name", "ds-agent_pool-test-acc"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},
			{
				Config: testAccScalrAgentPoolAccountDataSourceByIDAndNameConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_agent_pool.test", "scalr_agent_pool.test"),
					resource.TestCheckResourceAttrSet("data.scalr_agent_pool.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "name", "ds-agent_pool-test-acc"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},
		},
	})
}
func TestAccScalrAgentPoolDataSource_basic_env(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolEnvDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_agent_pool.test", "scalr_agent_pool.test"),
					resource.TestCheckResourceAttrSet("data.scalr_agent_pool.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "name", "agent_pool-test-env-ds"),
					resource.TestCheckResourceAttr("data.scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

var testAccScalrAgentPoolAccountDataSourceByIDConfig = fmt.Sprintf(`
resource scalr_environment test-new {
  name       = "test-env-new-for-pool"
  account_id = "%s"
}

resource "scalr_agent_pool" "test" {
  name       = "ds-agent_pool-test-acc"
  environments = [scalr_environment.test-new.id]
}

data "scalr_agent_pool" "test" {
  id         = scalr_agent_pool.test.id
  account_id = scalr_agent_pool.test.account_id
}`, defaultAccount)

var testAccScalrAgentPoolAccountDataSourceByNameConfig = fmt.Sprintf(`
resource "scalr_agent_pool" "test" {
  name       = "ds-agent_pool-test-acc"
  account_id = "%s"
}

data "scalr_agent_pool" "test" {
  name       = scalr_agent_pool.test.name
  account_id = scalr_agent_pool.test.account_id
}`, defaultAccount)

var testAccScalrAgentPoolAccountDataSourceByIDAndNameConfig = fmt.Sprintf(`
resource "scalr_agent_pool" "test" {
  name       = "ds-agent_pool-test-acc"
  account_id = "%s"
}

data "scalr_agent_pool" "test" {
  id         = scalr_agent_pool.test.id
  name       = scalr_agent_pool.test.name
  account_id = scalr_agent_pool.test.account_id
}`, defaultAccount)

func testAccScalrAgentPoolEnvDataSourceConfig() string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "agent_pool-test-123"
  account_id = "%s"

}
resource "scalr_agent_pool" "test" {
  name           = "agent_pool-test-env-ds"
  account_id     = "%s"
  environment_id = scalr_environment.test.id
}

data "scalr_agent_pool" "test" {
  name           = scalr_agent_pool.test.name
  account_id     = scalr_agent_pool.test.account_id
  environment_id = scalr_environment.test.id
}`, defaultAccount, defaultAccount)
}
