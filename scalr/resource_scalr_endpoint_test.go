package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccEndpoint_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "secret_key", "my-secret-key"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "timeout", "15"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "max_attempts", "3"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "url", "https://example.com/endpoint"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "environment_id", "existing-env"),
				),
			},
		},
	})
}

func TestAccEndpoint_update(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "secret_key", "my-secret-key"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "timeout", "15"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "max_attempts", "3"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "url", "https://example.com/endpoint"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "environment_id", "existing-env"),
				),
			},
			{
				Config: testAccEndpointConfigUpdate(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "timeout", "10"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "max_attempts", "5"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test-ep", "url", "https://example.com/endpoint-updated"),
				),
			},
		},
	})
}

func testAccEndpointConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_endpoint" "test-ep" {
  name         = "test endpoint-%d"
  secret_key   = "my-secret-key" 
  timeout      = 15               
  max_attempts = 3                
  url          = "https://example.com/endpoint"
  environment_id = "existing-env"
}`, rInt)
}

func testAccEndpointConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_endpoint" "test-ep" {
  name         = "test endpoint-%d"
  secret_key   = "my-secret-key" 
  timeout      = 10               
  max_attempts = 5                
  url          = "https://example.com/endpoint-updated"
  environment_id = "existing-env"
}`, rInt)
}
