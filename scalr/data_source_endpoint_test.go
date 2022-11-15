package scalr

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccEndpointDataSource_basic(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	rInt := rand.Intn(100)

	cutRInt := strconv.Itoa(rInt)[:len(strconv.Itoa(rInt))-1]

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
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
				Config:      testAccEndpointDataSourceNotFoundAlmostTheSameNameConfig(rInt, cutRInt),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Endpoint with name 'test endpoint-%s' not found", cutRInt)),
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
  account_id = scalr_environment.test.account_id
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

func testAccEndpointDataSourceNotFoundAlmostTheSameNameConfig(rInt int, cutRInt string) string {
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
  name           = "test endpoint-%[3]s"
}`, rInt, defaultAccount, cutRInt)
}
