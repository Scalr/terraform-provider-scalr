package provider

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
)

func TestAccCurrentAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories:  protoV5ProviderFactories(t),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					_ = os.Unsetenv(defaults.CurrentAccountIDEnvVar)
				},
				Config:      testAccCurrentAccountDataSourceConfig(),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Current account is not set"),
			},
			{
				PreConfig: func() {
					_ = os.Setenv(defaults.CurrentAccountIDEnvVar, defaultAccount)
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
