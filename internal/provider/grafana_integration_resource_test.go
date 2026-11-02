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

const testGrafanaEndpoint = "https://example.org/otlp"

func TestAccScalrGrafanaIntegrationResource_loki(t *testing.T) {
	endpoint := os.Getenv("TEST_GRAFANA_LOKI_ENDPOINT")
	token := os.Getenv("TEST_GRAFANA_LOKI_TOKEN")
	if endpoint == "" || token == "" {
		t.Skip("Please set TEST_GRAFANA_LOKI_ENDPOINT and TEST_GRAFANA_LOKI_TOKEN to run this test")
	}
	// A trailing slash is the same Loki, but a changed endpoint for the API.
	otherEndpoint := endpoint + "/"
	name := acctest.RandomWithPrefix("test-grafana")
	newName := acctest.RandomWithPrefix("test-grafana")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrGrafanaIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrGrafanaIntegrationLokiBearer(name, endpoint, token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrGrafanaIntegrationExists("scalr_grafana_integration.test"),
					resource.TestMatchResourceAttr("scalr_grafana_integration.test", "id", regexp.MustCompile(`^in-`)),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "name", name),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "kind", "loki"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "endpoint", endpoint),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "auth_type", "bearer"),
					resource.TestCheckNoResourceAttr("scalr_grafana_integration.test", "username"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "secret", token),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "tenant_id", "tenant-1"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "status", "active"),
					resource.TestCheckNoResourceAttr("scalr_grafana_integration.test", "err_message"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "account_id", defaultAccount),
				),
			},
			{
				Config: testAccScalrGrafanaIntegrationLokiBearer(newName, endpoint, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "name", newName),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "status", "active"),
				),
			},
			{
				// The API never returns the secret; make sure that does not show up as drift.
				Config: testAccScalrGrafanaIntegrationLokiBearer(newName, endpoint, token),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ResourceName:      "scalr_grafana_integration.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The API never returns the secret, so it cannot be imported.
				ImportStateVerifyIgnore: []string{"secret"},
			},
			{
				// The API drops the stored secret on an endpoint change unless it is sent again.
				Config: testAccScalrGrafanaIntegrationLokiBearer(newName, otherEndpoint, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "endpoint", otherEndpoint),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "auth_type", "bearer"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "secret", token),
				),
			},
		},
	})
}

func TestAccScalrGrafanaIntegrationResource_cloud(t *testing.T) {
	endpoint := os.Getenv("TEST_GRAFANA_ENDPOINT")
	instanceID := os.Getenv("TEST_GRAFANA_INSTANCE_ID")
	token := os.Getenv("TEST_GRAFANA_TOKEN")
	if endpoint == "" || instanceID == "" || token == "" {
		t.Skip("Please set TEST_GRAFANA_ENDPOINT, TEST_GRAFANA_INSTANCE_ID and TEST_GRAFANA_TOKEN to run this test")
	}

	name := acctest.RandomWithPrefix("test-grafana")
	newName := acctest.RandomWithPrefix("test-grafana")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrGrafanaIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrGrafanaIntegrationCloud(name, endpoint, instanceID, token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrGrafanaIntegrationExists("scalr_grafana_integration.test"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "name", name),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "kind", "cloud"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "endpoint", endpoint),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "auth_type", "basic"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "username", instanceID),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "secret", token),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "status", "active"),
					resource.TestCheckNoResourceAttr("scalr_grafana_integration.test", "err_message"),
				),
			},
			{
				Config: testAccScalrGrafanaIntegrationCloud(name, endpoint, instanceID, token),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: testAccScalrGrafanaIntegrationCloud(newName, endpoint, instanceID, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "name", newName),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "status", "active"),
				),
			},
			{
				ResourceName:            "scalr_grafana_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccScalrGrafanaIntegrationResource_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "tempo"`, `auth_type = "bearer"`, testGrafanaEndpoint, `secret = "x"`),
				ExpectError: regexp.MustCompile(`Attribute kind value must be one of`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "token"`, testGrafanaEndpoint, ""),
				ExpectError: regexp.MustCompile(`Attribute auth_type value must be one of`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "bearer"`, "http://example.com/otlp", `secret = "x"`),
				ExpectError: regexp.MustCompile(`URL scheme must be https`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "bearer"`, "https://example.com/otlp/v1/logs/", `secret = "x"`),
				ExpectError: regexp.MustCompile(`Enter the base OTLP URL without /v1/logs`),
			},
			{
				Config: testAccScalrGrafanaIntegrationRaw(
					`kind = "cloud"`, `auth_type = "bearer"`, testGrafanaEndpoint, `secret = "x"`,
				),
				ExpectError: regexp.MustCompile(`Grafana Cloud requires basic auth`),
			},
			{
				Config: testAccScalrGrafanaIntegrationRaw(
					`kind = "cloud"`, `auth_type = "basic"`, testGrafanaEndpoint,
					"username = \"1\"\n  secret = \"x\"\n  tenant_id = \"t\"",
				),
				ExpectError: regexp.MustCompile(`"tenant_id" is only used with kind "loki"`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "basic"`, testGrafanaEndpoint, `secret = "x"`),
				ExpectError: regexp.MustCompile(`"username" is required when "auth_type" is "basic"`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "bearer"`, testGrafanaEndpoint, ""),
				ExpectError: regexp.MustCompile(`"secret" is required when "auth_type" is "bearer"`),
			},
		},
	})
}

func testAccCheckScalrGrafanaIntegrationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		scalrClient := createScalrClientV2()

		_, err := scalrClient.GrafanaIntegration.GetGrafanaIntegration(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Grafana integration %s: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckScalrGrafanaIntegrationDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_grafana_integration" {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}
		_, err := scalrClient.GrafanaIntegration.GetGrafanaIntegration(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Grafana integration %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, client.ErrNotFound) {
			return err
		}
	}

	return nil
}

func testAccScalrGrafanaIntegrationLokiBearer(name, endpoint, token string) string {
	return fmt.Sprintf(`
resource "scalr_grafana_integration" "test" {
  name      = %q
  kind      = "loki"
  endpoint  = %q
  auth_type = "bearer"
  secret    = %q
  tenant_id = "tenant-1"
}`, name, endpoint, token)
}

func testAccScalrGrafanaIntegrationCloud(name, endpoint, instanceID, token string) string {
	return fmt.Sprintf(`
resource "scalr_grafana_integration" "test" {
  name      = %q
  kind      = "cloud"
  endpoint  = %q
  auth_type = "basic"
  username  = %q
  secret    = %q
}`, name, endpoint, instanceID, token)
}

func testAccScalrGrafanaIntegrationRaw(kind, authType, endpoint, extra string) string {
	return fmt.Sprintf(`
resource "scalr_grafana_integration" "test" {
  name     = "test-grafana-invalid"
  endpoint = %q
  %s
  %s
  %s
}`, endpoint, kind, authType, extra)
}
