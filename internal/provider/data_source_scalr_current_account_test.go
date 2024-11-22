package provider

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCurrentAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					_ = os.Unsetenv(currentAccountIDEnvVar)
				},
				Config:      testAccCurrentAccountDataSourceConfig(),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Current account is not set"),
			},
			{
				PreConfig: func() {
					_ = os.Setenv(currentAccountIDEnvVar, defaultAccount)
				},
				Config:   testAccCurrentAccountDataSourceConfig(),
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_current_account.test", "id", defaultAccount),
					resource.TestCheckResourceAttr(
						"data.scalr_current_account.test", "name", "mainiacp"),
				),
			},
		},
	})
}

func testAccCurrentAccountDataSourceConfig() string {
	return "data scalr_current_account test {}"
}
