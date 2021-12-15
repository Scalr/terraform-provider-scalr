package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccEnvironmentDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_environment.test", "name", fmt.Sprintf("test-env-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "status", "Active"),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "cost_estimation_enabled", "false"),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "cloud_credentials.%", "0"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.username"),
				),
			},
			{
				Config: testAccEnvironmentDataSourceAccesByNmaeConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_environment.test", "name", fmt.Sprintf("test-env-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "status", "Active"),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "cost_estimation_enabled", "false"),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "cloud_credentials.%", "0"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.username"),
				),
			},
			{
				Config:      testAccEnvironmentDataSourceNotFoundConfig(),
				ExpectError: regexp.MustCompile("Environment with ID 'env-123' not found or user unauthorized"),
				PlanOnly:    true,
			},
			{
				Config:      testAccEnvironmentDataSourceNotFoundByNameConfig(),
				ExpectError: regexp.MustCompile("Environment with name 'env-foo-bar-baz' not found or user unauthorized"),
				PlanOnly:    true,
			},
			{
				Config:      testAccEnvironmentNoNameNitherIdSetConfig(),
				ExpectError: regexp.MustCompile("At least one argument 'id' or 'name' is required, but no definitions was found"),
				PlanOnly:    true,
			},
		},
	})
}

func testAccEnvironmentDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "%s"
}

data "scalr_environment" "test" {
  id         = scalr_environment.test.id
}`, rInt, defaultAccount)
}

func testAccEnvironmentDataSourceAccesByNmaeConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "%s"
}

data "scalr_environment" "test" {
  name         = scalr_environment.test.name
  account_id = "%s"
}`, rInt, defaultAccount, defaultAccount)
}

func testAccEnvironmentDataSourceNotFoundConfig() string {
	return `
data "scalr_environment" "test" {
  id = "env-123"
}`
}

func testAccEnvironmentDataSourceNotFoundByNameConfig() string {
	return `
data "scalr_environment" "test" {
  name = "env-foo-bar-baz"
}`
}

func testAccEnvironmentNoNameNitherIdSetConfig() string {
	return `data "scalr_environment" "test" {}`
}
