package scalr

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccEnvironmentDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

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
					resource.TestCheckResourceAttr("data.scalr_environment.test", "account_id", DefaultAccount),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "account_id", DefaultAccount),
					resource.TestCheckResourceAttr("data.scalr_environment.test", "cloud_credentials.%", "0"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("data.scalr_environment.test", "created_by.0.username"),
				),
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
}`, rInt, DefaultAccount)
}
