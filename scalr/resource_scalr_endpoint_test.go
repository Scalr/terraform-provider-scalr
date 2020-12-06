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
	secretKey := "stong_key_with_UPPERCASE_letter_at_leaast_1_number"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig(rInt, secretKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "secret_key", secretKey),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "timeout", "15"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "max_attempts", "3"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "url", "https://example.com/endpoint"),
					resource.TestCheckResourceAttrSet(
						"scalr_endpoint.test", "environment_id"),
				),
			},
		},
	})
}

func TestAccEndpoint_update(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	secretKey := "stong_key_with_UPPERCASE_letter_at_leaast_1_number"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig(rInt, secretKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "secret_key", secretKey),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "timeout", "15"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "max_attempts", "3"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "url", "https://example.com/endpoint"),
					resource.TestCheckResourceAttrSet(
						"scalr_endpoint.test", "environment_id"),
				),
			},
			{
				Config: testAccEndpointConfigUpdate(rInt, secretKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "name", fmt.Sprintf("test endpoint-%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "timeout", "10"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "max_attempts", "5"),
					resource.TestCheckResourceAttr(
						"scalr_endpoint.test", "url", "https://example.com/endpoint-updated"),
				),
			},
		},
	})
}

func testAccEndpointConfig(rInt int, secretKey string) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "acc-svrcncgh453bi8g"
}

resource scalr_endpoint test {
  name         = "test endpoint-%[1]d"
  secret_key   = "%s"
  timeout      = 15               
  max_attempts = 3                
  url          = "https://example.com/endpoint"
  environment_id = scalr_environment.test.id
}`, rInt, secretKey)
}

func testAccEndpointConfigUpdate(rInt int, secretKey string) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "acc-svrcncgh453bi8g"
}

resource scalr_endpoint test {
  name         = "test endpoint-%[1]d"
  secret_key   = "%s"
  timeout      = 10               
  max_attempts = 5                
  url          = "https://example.com/endpoint-updated"
  environment_id = scalr_environment.test.id
}`, rInt, secretKey)
}
