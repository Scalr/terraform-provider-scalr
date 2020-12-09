package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccEndpointDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

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
}`, rInt, DefaultAccount)
}
