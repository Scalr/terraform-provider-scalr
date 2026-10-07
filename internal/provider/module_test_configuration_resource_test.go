package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrModuleTestConfiguration_basic(t *testing.T) {
	resourceName := "scalr_module_test_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestConfigurationBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "module_id", "scalr_module.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "failure_behavior", "notify"),
					resource.TestCheckResourceAttr(resourceName, "trigger_on_pr_activity_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "trigger_on_new_version_enabled", "false"),
				),
			},
		},
	})
}

func TestAccScalrModuleTestConfiguration_update(t *testing.T) {
	resourceName := "scalr_module_test_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestConfigurationBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "failure_behavior", "notify"),
				),
			},
			{
				Config: testAccScalrModuleTestConfigurationUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "failure_behavior", "failure"),
					resource.TestCheckResourceAttr(resourceName, "trigger_on_pr_activity_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "trigger_on_new_version_enabled", "true"),
				),
			},
		},
	})
}

func TestAccScalrModuleTestConfiguration_import(t *testing.T) {
	resourceName := "scalr_module_test_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestConfigurationBasic(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["scalr_module.test"]
					if !ok {
						return "", fmt.Errorf("module resource not found in state")
					}
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func testAccCheckScalrModuleTestConfigurationDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_module_test_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// There is no delete endpoint: destroying the resource disables it instead.
		tc, err := scalrClient.ModuleTestConfigurations.Read(ctx, rs.Primary.ID)
		if err != nil {
			continue
		}
		if tc.Enabled {
			return fmt.Errorf("Module test configuration %s is still enabled", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrModuleTestConfigurationBasic() string {
	return testAccScalrModule() + `
resource "scalr_module_test_configuration" "test" {
  module_id = scalr_module.test.id
}
`
}

func testAccScalrModuleTestConfigurationUpdated() string {
	return testAccScalrModule() + `
resource "scalr_module_test_configuration" "test" {
  module_id                      = scalr_module.test.id
  enabled                        = false
  failure_behavior               = "failure"
  trigger_on_pr_activity_enabled = true
  trigger_on_new_version_enabled = true
}
`
}
