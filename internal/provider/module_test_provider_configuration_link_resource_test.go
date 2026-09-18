package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrModuleTestProviderConfigurationLink_basic(t *testing.T) {
	resourceName := "scalr_module_test_provider_configuration_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestProviderConfigurationLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestProviderConfigurationLinkBasic(),
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

func TestAccScalrModuleTestProviderConfigurationLink_update(t *testing.T) {
	resourceName := "scalr_module_test_provider_configuration_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestProviderConfigurationLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestProviderConfigurationLinkBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "provider_configuration_id", "scalr_provider_configuration.test1", "id",
					),
				),
			},
			{
				Config: testAccScalrModuleTestProviderConfigurationLinkUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "provider_configuration_id", "scalr_provider_configuration.test2", "id",
					),
				),
			},
		},
	})
}

func TestAccScalrModuleTestProviderConfigurationLink_import(t *testing.T) {
	resourceName := "scalr_module_test_provider_configuration_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testVcsAccGithubTokenPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleTestProviderConfigurationLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleTestProviderConfigurationLinkBasic(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrModuleTestProviderConfigurationLinkDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_module_test_provider_configuration_link" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.ModuleTestProviderConfigurationLinks.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Module test provider configuration link %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrModuleTestProviderConfigurationLinkBasic() string {
	return testAccScalrModule() + fmt.Sprintf(`
resource "scalr_module_test_configuration" "test" {
  module_id = scalr_module.test.id
}

resource "scalr_provider_configuration" "test1" {
  name                      = "test-mtpcl-1"
  account_id                = "%[1]s"
  is_allowed_in_module_test = true
  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    access_key       = "access_key"
    secret_key       = "secret_key"
  }
}

resource "scalr_provider_configuration" "test2" {
  name                      = "test-mtpcl-2"
  account_id                = "%[1]s"
  is_allowed_in_module_test = true
  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    access_key       = "access_key"
    secret_key       = "secret_key"
  }
}

resource "scalr_module_test_provider_configuration_link" "test" {
  test_configuration_id     = scalr_module_test_configuration.test.id
  provider_configuration_id = scalr_provider_configuration.test1.id
}
`, defaultAccount)
}

func testAccScalrModuleTestProviderConfigurationLinkUpdated() string {
	return testAccScalrModule() + fmt.Sprintf(`
resource "scalr_module_test_configuration" "test" {
  module_id = scalr_module.test.id
}

resource "scalr_provider_configuration" "test1" {
  name                      = "test-mtpcl-1"
  account_id                = "%[1]s"
  is_allowed_in_module_test = true
  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    access_key       = "access_key"
    secret_key       = "secret_key"
  }
}

resource "scalr_provider_configuration" "test2" {
  name                      = "test-mtpcl-2"
  account_id                = "%[1]s"
  is_allowed_in_module_test = true
  aws {
    account_type     = "regular"
    credentials_type = "access_keys"
    access_key       = "access_key"
    secret_key       = "secret_key"
  }
}

resource "scalr_module_test_provider_configuration_link" "test" {
  test_configuration_id     = scalr_module_test_configuration.test.id
  provider_configuration_id = scalr_provider_configuration.test2.id
}
`, defaultAccount)
}
