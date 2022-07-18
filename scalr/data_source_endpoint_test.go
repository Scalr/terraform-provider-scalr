package scalr

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccEndpointDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()
	for {
		if rInt >= 100 {
			break
		}
		rInt = GetRandomInteger()
	}

	cuttedRInt := strconv.Itoa(rInt)[:len(strconv.Itoa(rInt))-1]

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_endpoint.test", "secret_key"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test", "timeout", "15"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test", "max_attempts", "3"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test", "url", "https://example.com/endpoint"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_endpoint.test", "environment_id"),
				),
			},
			{
				Config: testAccEndpointDataSourceAccessByNameConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_endpoint.test", "secret_key"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test", "timeout", "15"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test", "max_attempts", "3"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test", "url", "https://example.com/endpoint"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_endpoint.test", "environment_id"),
				),
			},
			{
				Config:      testAccEndpointDataSourceNotFoundAlmostTheSameNameConfig(rInt, cuttedRInt),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Endpoint with name 'test endpoint-%s' not found", cuttedRInt)),
				PlanOnly:    true,
			},
			{
				Config:      testAccEndpointDataSourceNotFoundByNameConfig(),
				ExpectError: regexp.MustCompile("Endpoint with name 'endpoint-foo-bar-baz' not found or user unauthorized"),
				PlanOnly:    true,
			},
			{
				Config:      testAccEndpointNeitherNameNorIdSetConfig(),
				ExpectError: regexp.MustCompile("At least one argument 'id' or 'name' is required, but no definitions was found"),
				PlanOnly:    true,
			},
			{
				Config:      testAccEndpointBothNameAndIdSetConfig(),
				ExpectError: regexp.MustCompile("Attributes 'name' and 'id' can not be set at the same time"),
				PlanOnly:    true,
			},
		},
	})
}

func testAccEndpointDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_endpoint test {
  name         = "test endpoint-%[1]d"
  timeout      = 15
  url          = "https://example.com/endpoint"
  environment_id = scalr_environment.test.id
}

data scalr_endpoint test {
  id         = scalr_endpoint.test.id
}`, rInt, defaultAccount)
}

func testAccEndpointDataSourceAccessByNameConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_endpoint test {
  name         = "test endpoint-%[1]d"
  timeout      = 15
  url          = "https://example.com/endpoint"
  environment_id = scalr_environment.test.id
}

data scalr_endpoint test {
  name = scalr_endpoint.test.name
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount)
}

func testAccEndpointDataSourceNotFoundByNameConfig() string {
	return `
data scalr_endpoint test {
  name = "endpoint-foo-bar-baz"
}`
}

func testAccEndpointNeitherNameNorIdSetConfig() string {
	return `data scalr_endpoint test {}`
}

func testAccEndpointBothNameAndIdSetConfig() string {
	return `data scalr_endpoint test {
		id = "foo"
		name = "bar"
	}`
}

func testAccEndpointDataSourceNotFoundAlmostTheSameNameConfig(rInt int, cuttedRInt string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_endpoint test {
  name         = "test endpoint-%[1]d"
  timeout      = 15
  url          = "https://example.com/endpoint"
  environment_id = scalr_environment.test.id
}

data scalr_endpoint test {
  name           = "test endpoint-%s"
}`, rInt, defaultAccount, cuttedRInt)
}
