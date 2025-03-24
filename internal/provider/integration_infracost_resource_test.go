package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
	"os"
	"testing"
)

func TestIntegrationInfracostResource_Create(t *testing.T) {
	apiKey := os.Getenv("TEST_INFRACOST_API_KEY")
	if len(apiKey) == 0 {
		t.Skip("Please set TEST_INFRACOST_API_KEY to run this test.")
	}
	var infracostIntegration scalr.InfracostIntegration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckInfracostIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIntegrationInfracostConfig(apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_integration_infracost.test", "id"),
					resource.TestCheckResourceAttr("scalr_integration_infracost.test", "name", "test-create"),
					testAccCheckInfracostIntegrationIsShared("scalr_integration_infracost.test", true, &infracostIntegration),
				),
			},
		},
	})
}

func TestIntegrationInfracostResource_Update(t *testing.T) {
	apiKey := os.Getenv("TEST_INFRACOST_API_KEY")
	if len(apiKey) == 0 {
		t.Skip("Please set TEST_INFRACOST_API_KEY to run this test.")
	}
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	var infracostIntegration scalr.InfracostIntegration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckInfracostIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIntegrationInfracostConfig(apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_integration_infracost.test", "name", "test-create"),
					testAccCheckInfracostIntegrationIsShared("scalr_integration_infracost.test", true, &infracostIntegration),
				),
			},
			{
				Config: testAccScalrIntegrationInfracostConfigUpdate(rNewName, apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_integration_infracost.test", "name", rNewName),
					testAccCheckInfracostIntegrationIsShared("scalr_integration_infracost.test", false, &infracostIntegration),
				),
			},
		},
	})
}

func TestIntegrationInfracostResource_ImportState(t *testing.T) {
	apiKey := os.Getenv("TEST_INFRACOST_API_KEY")
	if len(apiKey) == 0 {
		t.Skip("Please set TEST_INFRACOST_API_KEY to run this test.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckInfracostIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrIntegrationInfracostConfig(apiKey),
			},
			{
				ResourceName:            "scalr_integration_infracost.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccScalrIntegrationInfracostConfig(apiKey string) string {
	return fmt.Sprintf(`
resource "scalr_integration_infracost" "test" {
  name         = "test-create"
  api_key      = "%s"
  environments = ["*"]
}`, apiKey)
}

func testAccScalrIntegrationInfracostConfigUpdate(name string, apiKey string) string {
	return fmt.Sprintf(`
resource "scalr_integration_infracost" "test" {
  name         = "%s"
  api_key      = "%s"
  environments = []
}`, name, apiKey)
}

func testAccCheckInfracostIntegrationDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_integration_infracost" {
			continue
		}

		_, err := scalrClient.InfracostIntegrations.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Infracost integration (%s) still exists", rs.Primary.ID)
		}

		if !errors.Is(err, scalr.ErrResourceNotFound) {
			return err
		}
	}

	return nil
}

func testAccCheckInfracostIntegrationIsShared(resourceName string, expectedIsShared bool, InfracostIntegration *scalr.InfracostIntegration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)
		readKey, err := scalrClient.InfracostIntegrations.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error reading Infracost Integration: %s", err)
		}

		*InfracostIntegration = *readKey

		if readKey.IsShared != expectedIsShared {
			return fmt.Errorf("Expected IsShared to be %t, but got %t", expectedIsShared, readKey.IsShared)
		}
		return nil
	}
}
