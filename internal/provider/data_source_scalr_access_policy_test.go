package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrAccessPolicyDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAccessPolicyDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_access_policy.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_access_policy.test", "subject.0.type", "user"),
					resource.TestCheckResourceAttr("data.scalr_access_policy.test", "subject.0.id", testUser),
					resource.TestCheckResourceAttr("data.scalr_access_policy.test", "is_system", "false"),
					resource.TestCheckResourceAttr("data.scalr_access_policy.test", "scope.0.type", "environment"),
					resource.TestCheckResourceAttr("data.scalr_access_policy.test", "role_ids.0", readOnlyRole),
					resource.TestCheckResourceAttr("data.scalr_access_policy.test", "role_ids.#", "1"),
				),
			},
			{
				Config:      testAccAccessPolicyDataSourceNotFoundConfig(),
				ExpectError: regexp.MustCompile("AccessPolicy 'ap-123' not found"),
				PlanOnly:    true,
			},
		},
	})
}
func testAccScalrAccessPolicyDataSourceConfig() string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name = "test-access-policies-provider-data-source"
  account_id = "%s"
}


resource "scalr_access_policy" "test" {
  subject {
    type = "user"
    id = "%s"
  }
  scope {
    type = "environment"
    id = scalr_environment.test.id
  }
  role_ids = [
    "%s"
  ]
}

data "scalr_access_policy" "test" {
   id = scalr_access_policy.test.id
}`, defaultAccount, testUser, readOnlyRole)
}

func testAccAccessPolicyDataSourceNotFoundConfig() string {
	return `
data "scalr_access_policy" "test" {
  id = "ap-123"
}`
}
