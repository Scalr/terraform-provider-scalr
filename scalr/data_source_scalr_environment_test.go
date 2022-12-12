package scalr

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccEnvironmentDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()
	for {
		if rInt >= 100 {
			break
		}
		rInt = GetRandomInteger()
	}

	cuttedRInt := strconv.Itoa(rInt)[:len(strconv.Itoa(rInt))-1]

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEnvironmentNeitherNameNorIdSetConfig(),
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      testAccEnvironmentBothNameAndIdSetConfig(),
				ExpectError: regexp.MustCompile("\"name\": conflicts with id"),
				PlanOnly:    true,
			},
			{
				Config: testAccEnvironmentDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scalr_environment.test", "name", fmt.Sprintf("test-env-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "status", "Active"),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "cost_estimation_enabled", "false"),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "cloud_credentials.%", "0"),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "tags.#", "0"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.username"),
				),
			},
			{
				Config: testAccEnvironmentDataSourceAccessByNameConfig(rInt),
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
				ExpectError: regexp.MustCompile("Environment 'env-123' not found"),
				PlanOnly:    true,
			},
			{
				Config:      testAccEnvironmentDataSourceNotFoundAlmostTheSameNameConfig(rInt, cuttedRInt),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Environment with name 'test-env-%s' not found", cuttedRInt)),
				PlanOnly:    true,
			},
			{
				Config:      testAccEnvironmentDataSourceNotFoundByNameConfig(),
				ExpectError: regexp.MustCompile("Environment with name 'env-foo-bar-baz' not found or user unauthorized"),
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

func testAccEnvironmentDataSourceAccessByNameConfig(rInt int) string {
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

func testAccEnvironmentNeitherNameNorIdSetConfig() string {
	return `data "scalr_environment" "test" {}`
}

func testAccEnvironmentBothNameAndIdSetConfig() string {
	return `data "scalr_environment" "test" {
		id = "foo"
		name = "bar"
	}`
}

func testAccEnvironmentDataSourceNotFoundAlmostTheSameNameConfig(rInt int, cuttedRInt string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "%s"
}

data "scalr_environment" "test" {
  name         = "test-env-%s"
}`, rInt, defaultAccount, cuttedRInt)
}
