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
						"data.scalr_endpoint.test-ep", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test-ep", "secret_key", "my-secret-key"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test-ep", "timeout", "15"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test-ep", "max_attempts", "3"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test-ep", "url", "https://example.com/endpoint"),
					resource.TestCheckResourceAttr(
						"data.scalr_endpoint.test-ep", "environment_id", "existing-env"),
				),
			},
		},
	})
}

func testAccEndpointDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_endpoint" "test-ep" {
  name         = "test endpoint-%d"
  secret_key   = "my-secret-key" 
  timeout      = 15
  url          = "https://example.com/endpoint"
  environment_id = "existing-env"
}

data "scalr_endpoint" "test-ep" {
  id         = "${scalr_endpoint.test-ep.id}"
}`, rInt)
}
