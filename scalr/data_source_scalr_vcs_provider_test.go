package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalrVcsProviderDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testVcsAccGithubTokenPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVcsProviderDataSourceConfigAllFilters(rInt, GITHUB_TOKEN),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_vcs_provider.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "name", fmt.Sprintf("vcs-provider-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "vcs_type", "github"),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "url", "https://github.com"),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "account_id", defaultAccount),
				),
			},
			{
				Config: testAccScalrVcsProviderDataSourceConfigFilterByName(rInt, GITHUB_TOKEN),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_vcs_provider.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "name", fmt.Sprintf("vcs-provider-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "vcs_type", "github"),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "url", "https://github.com"),
					resource.TestCheckResourceAttr(
						"data.scalr_vcs_provider.test", "account_id", defaultAccount),
				),
			},
			{
				Config: `
				data scalr_vcs_provider test {
				  vcs_type = "github"
				}`,
				ExpectError: regexp.MustCompile("Your query returned more than one result. Please try a more specific search criteria"),
				PlanOnly:    true,
			},
			{
				Config: `
				data scalr_vcs_provider test {
				  name = "not-existing-vcs"
				}`,
				ExpectError: regexp.MustCompile("Could not find vcs provider matching you query"),
				PlanOnly:    true,
			},
		},
	})
}

func testAccScalrVcsProviderDataSourceConfigAllFilters(rInt int, token string) string {
	return fmt.Sprintf(`
resource scalr_vcs_provider test {
  name       = "vcs-provider-test-%[1]d"
  vcs_type   = "github"
  token      = "%s"
  account_id = "%s"
}

data scalr_vcs_provider test {
  name     = scalr_vcs_provider.test.name
  vcs_type = scalr_vcs_provider.test.vcs_type
  account_id  = scalr_vcs_provider.test.account_id
}`, rInt, token, defaultAccount)
}

func testAccScalrVcsProviderDataSourceConfigFilterByName(rInt int, token string) string {
	return fmt.Sprintf(`
resource scalr_vcs_provider test {
  name        = "vcs-provider-test-%[1]d"
  vcs_type    = "github"
  token       = "%s"
  account_id  = "%s"
}

data scalr_vcs_provider test {
  name     = scalr_vcs_provider.test.name
}`, rInt, token, defaultAccount)
}
