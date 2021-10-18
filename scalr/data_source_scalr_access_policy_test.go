package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalrAccessPolicyDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
				ExpectError: regexp.MustCompile("IamAccessPolicy with ID 'ap-123' not found or user unauthorized"),
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
