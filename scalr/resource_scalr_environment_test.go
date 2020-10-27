package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccEnvironment_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_environment.test", "name", fmt.Sprintf("test-env-%d", rInt)),
					resource.TestCheckResourceAttr("scalr_environment.test", "cost_estimation_enabled", "true"),
					resource.TestCheckResourceAttr("scalr_environment.test", "status", "Active"),
					resource.TestCheckResourceAttr("scalr_environment.test", "account_id", "acc-svrcncgh453bi8g"),
					resource.TestCheckResourceAttr("scalr_environment.test", "cloud_credentials.%", "0"),
					resource.TestCheckResourceAttr("scalr_environment.test", "policy_groups.%", "0"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("scalr_environment.test", "created_by.0.username"),
				),
			},
		},
	})
}

// func TestAccEndpoint_update(t *testing.T) {
// 	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:  func() { testAccPreCheck(t) },
// 		Providers: testAccProviders,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccEndpointConfig(rInt),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "name", fmt.Sprintf("test endpoint-%d", rInt)),
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "secret_key", "my-secret-key"),
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "timeout", "15"),
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "max_attempts", "3"),
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "url", "https://example.com/endpoint"),
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "environment_id", "existing-env"),
// 				),
// 			},
// 			{
// 				Config: testAccEndpointConfigUpdate(rInt),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "name", fmt.Sprintf("test endpoint-%d", rInt)),
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "timeout", "10"),
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "max_attempts", "5"),
// 					resource.TestCheckResourceAttr(
// 						"scalr_endpoint.test-ep", "url", "https://example.com/endpoint-updated"),
// 				),
// 			},
// 		},
// 	})
// }

func testAccCheckScalrEnvironmentDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_environment" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Environments.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Environment %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccEnvironmentConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "acc-svrcncgh453bi8g"
  cost_estimation_enabled = true

}`, rInt)
}

// func testAccEndpointEnvironmentUpdate(rInt int) string {
// 	return fmt.Sprintf(`
// resource "scalr_" "test-ep" {
//   name         = "test endpoint-%d"
//   secret_key   = "my-secret-key"
//   timeout      = 10
//   max_attempts = 5
//   url          = "https://example.com/endpoint-updated"
//   environment_id = "existing-env"
// }`, rInt)
// }
