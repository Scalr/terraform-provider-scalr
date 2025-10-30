package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCurrentEnvironment_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories:  protoV5ProviderFactories(t),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					_ = os.Unsetenv(CurrentEnvironmentIdEnvVar)
				},
				Config:      testCurrentEnvironment_basicDataSourceConfig(),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(fmt.Sprintf("Current environmnet is not set. `%s` OS environment variable must be set", CurrentEnvironmentIdEnvVar)),
			},
			{
				PreConfig: func() {
					_ = os.Setenv(CurrentEnvironmentIdEnvVar, "env-123")
				},
				Config:   testCurrentEnvironment_basicDataSourceConfig(),
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_current_environment.test", "id", "env-123"),
				),
			},
		},
	})
}

func testCurrentEnvironment_basicDataSourceConfig() string {
	return "data scalr_current_environment test {}"
}
