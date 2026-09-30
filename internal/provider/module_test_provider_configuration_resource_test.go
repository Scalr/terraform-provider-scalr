package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrModuleTestProviderConfiguration_basic(t *testing.T) {
	resourceName := "scalr_module_test_provider_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestProviderConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestProviderConfigurationBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(
						resourceName, "test_configuration_id", "scalr_module_test_configuration.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "provider_configuration_id", "scalr_provider_configuration.test1", "id",
					),
				),
			},
		},
	})
}

func TestAccScalrModuleTestProviderConfiguration_update(t *testing.T) {
	resourceName := "scalr_module_test_provider_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestProviderConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestProviderConfigurationBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "provider_configuration_id", "scalr_provider_configuration.test1", "id",
					),
				),
			},
			{
				Config: testAccScalrModuleTestProviderConfigurationUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "provider_configuration_id", "scalr_provider_configuration.test2", "id",
					),
				),
			},
		},
	})
}

func TestAccScalrModuleTestProviderConfiguration_import(t *testing.T) {
	resourceName := "scalr_module_test_provider_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestProviderConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestProviderConfigurationBasic(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrModuleTestProviderConfigurationDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_module_test_provider_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.ModuleTestProviderConfigurationLinks.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Module test provider configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrModuleTestProviderConfigurationBasic() string {
	return testAccScalrModule() + fmt.Sprintf(`
resource "scalr_module_test_configuration" "test" {
  module_id = scalr_module.test.id
}

resource "scalr_provider_configuration" "test1" {
  name                      = "test-mtpcl-1"
  account_id                = "%[1]s"
  is_allowed_in_module_test = true
  custom {
    provider_name = "consul"
    argument {
      name  = "address"
      value = "127.0.0.1:8500"
    }
  }
}

resource "scalr_provider_configuration" "test2" {
  name                      = "test-mtpcl-2"
  account_id                = "%[1]s"
  is_allowed_in_module_test = true
  custom {
    provider_name = "consul"
    argument {
      name  = "address"
      value = "127.0.0.1:8500"
    }
  }
}

resource "scalr_module_test_provider_configuration" "test" {
  test_configuration_id     = scalr_module_test_configuration.test.id
  provider_configuration_id = scalr_provider_configuration.test1.id
}
`, defaultAccount)
}

func testAccScalrModuleTestProviderConfigurationUpdated() string {
	return testAccScalrModule() + fmt.Sprintf(`
resource "scalr_module_test_configuration" "test" {
  module_id = scalr_module.test.id
}

resource "scalr_provider_configuration" "test1" {
  name                      = "test-mtpcl-1"
  account_id                = "%[1]s"
  is_allowed_in_module_test = true
  custom {
    provider_name = "consul"
    argument {
      name  = "address"
      value = "127.0.0.1:8500"
    }
  }
}

resource "scalr_provider_configuration" "test2" {
  name                      = "test-mtpcl-2"
  account_id                = "%[1]s"
  is_allowed_in_module_test = true
  custom {
    provider_name = "consul"
    argument {
      name  = "address"
      value = "127.0.0.1:8500"
    }
  }
}

resource "scalr_module_test_provider_configuration" "test" {
  test_configuration_id     = scalr_module_test_configuration.test.id
  provider_configuration_id = scalr_provider_configuration.test2.id
}
`, defaultAccount)
}
