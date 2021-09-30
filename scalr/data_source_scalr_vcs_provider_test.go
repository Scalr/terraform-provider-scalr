package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalrVcsProviderDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVcsProviderDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_vcs_provider.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "name", fmt.Sprintf("vcs-provider-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "vcs_type", "github"),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "account", defaultAccount),
				),
			},
		},
	})
}

func testAccScalrVcsProviderDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_vcs_provider test {
  name       = "vcs-provider-test-%[1]d"
  vcs_type   = "github"
  token      = "%s"
  account_id = "%s"
}

data scalr_vcs_provider test {
  id       = scalr_vcs_provider.test.id
  name     = scalr_vcs_provider.test.name
  vcs_type = scalr_vcs_provider.test.vcs_type
  account  = scalr_environment.test.account
}`, rInt, GITHUB_TOKEN, defaultAccount)
}
