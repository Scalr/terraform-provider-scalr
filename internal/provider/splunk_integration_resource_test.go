package provider

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/scalr/go-scalr/v2/scalr/client"
)

func TestAccScalrSplunkIntegrationResource_basic(t *testing.T) {
	endpoint := os.Getenv("TEST_SPLUNK_ENDPOINT")
	token := os.Getenv("TEST_SPLUNK_TOKEN")
	index := os.Getenv("TEST_SPLUNK_INDEX")
	if endpoint == "" || token == "" || index == "" {
		t.Skip("Please set TEST_SPLUNK_ENDPOINT, TEST_SPLUNK_TOKEN and TEST_SPLUNK_INDEX to run this test")
	}
	// A trailing slash is the same HEC, but a changed endpoint for the API.
	otherEndpoint := endpoint + "/"
	name := acctest.RandomWithPrefix("test-splunk")
	newName := acctest.RandomWithPrefix("test-splunk")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrSplunkIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrSplunkIntegrationConfig(name, endpoint, token, index),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrSplunkIntegrationExists("scalr_splunk_integration.test"),
					resource.TestMatchResourceAttr("scalr_splunk_integration.test", "id", regexp.MustCompile(`^in-`)),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "name", name),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "endpoint", endpoint),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "token", token),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "index", index),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "status", "active"),
					resource.TestCheckNoResourceAttr("scalr_splunk_integration.test", "err_message"),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "account_id", defaultAccount),
				),
			},
			{
				// The API never returns the token; make sure that does not show up as drift.
				Config: testAccScalrSplunkIntegrationConfig(name, endpoint, token, index),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				// Removing the index clears it, so events go to the default index of the token.
				Config: testAccScalrSplunkIntegrationConfigWithoutIndex(newName, endpoint, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "name", newName),
					resource.TestCheckNoResourceAttr("scalr_splunk_integration.test", "index"),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "status", "active"),
				),
			},
			{
				ResourceName:      "scalr_splunk_integration.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The API never returns the token, so it cannot be imported.
				ImportStateVerifyIgnore: []string{"token"},
			},
			{
				// The API drops the stored token on an endpoint change unless it is sent again.
				Config: testAccScalrSplunkIntegrationConfig(newName, otherEndpoint, token, index),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "endpoint", otherEndpoint),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "token", token),
					resource.TestCheckResourceAttr("scalr_splunk_integration.test", "index", index),
				),
			},
		},
	})
}

func TestAccScalrSplunkIntegrationResource_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrSplunkIntegrationConfig("test-splunk-invalid", "http://splunk.example.com:8088", "x", "main"),
				ExpectError: regexp.MustCompile(`URL scheme must be https`),
			},
			{
				Config:      testAccScalrSplunkIntegrationConfig("test-splunk-invalid", "https://splunk.example.com:8088/services/collector/event", "x", "main"),
				ExpectError: regexp.MustCompile(`Enter the base HEC URL without /services/collector`),
			},
			{
				Config:      testAccScalrSplunkIntegrationConfig("test-splunk-invalid", "https://splunk.example.com:8088", "x", " "),
				ExpectError: regexp.MustCompile(`Attribute index must not be empty`),
			},
		},
	})
}

func testAccCheckScalrSplunkIntegrationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		scalrClient := createScalrClientV2()

		_, err := scalrClient.SplunkIntegration.GetSplunkIntegration(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Splunk integration %s: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckScalrSplunkIntegrationDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_splunk_integration" {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}
		_, err := scalrClient.SplunkIntegration.GetSplunkIntegration(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Splunk integration %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, client.ErrNotFound) {
			return err
		}
	}

	return nil
}

func testAccScalrSplunkIntegrationConfig(name, endpoint, token, index string) string {
	return fmt.Sprintf(`
resource "scalr_splunk_integration" "test" {
  name     = %q
  endpoint = %q
  token    = %q
  index    = %q
}`, name, endpoint, token, index)
}

func testAccScalrSplunkIntegrationConfigWithoutIndex(name, endpoint, token string) string {
	return fmt.Sprintf(`
resource "scalr_splunk_integration" "test" {
  name     = %q
  endpoint = %q
  token    = %q
}`, name, endpoint, token)
}
