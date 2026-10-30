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

// The Loki tests point at an endpoint that rejects the connection check,
// so the integration is expected to end up in the `failed` status.
const (
	testGrafanaFailingEndpoint      = "https://example.com/otlp"
	testGrafanaOtherFailingEndpoint = "https://example.org/otlp"
)

func TestAccScalrGrafanaIntegrationResource_loki(t *testing.T) {
	name := acctest.RandomWithPrefix("test-grafana")
	newName := acctest.RandomWithPrefix("test-grafana")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrGrafanaIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrGrafanaIntegrationLokiNone(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrGrafanaIntegrationExists("scalr_grafana_integration.test"),
					resource.TestMatchResourceAttr("scalr_grafana_integration.test", "id", regexp.MustCompile(`^in-`)),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "name", name),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "kind", "loki"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "endpoint", testGrafanaFailingEndpoint),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "auth_type", "none"),
					resource.TestCheckNoResourceAttr("scalr_grafana_integration.test", "username"),
					resource.TestCheckNoResourceAttr("scalr_grafana_integration.test", "secret"),
					resource.TestCheckNoResourceAttr("scalr_grafana_integration.test", "tenant_id"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "status", "failed"),
					resource.TestCheckResourceAttrSet("scalr_grafana_integration.test", "err_message"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "account_id", defaultAccount),
				),
			},
			{
				// A server-set `failed` status does not cause drift.
				Config: testAccScalrGrafanaIntegrationLokiNone(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: testAccScalrGrafanaIntegrationLokiBearer(newName, testGrafanaFailingEndpoint),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "name", newName),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "auth_type", "bearer"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "secret", "fake-token"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "tenant_id", "tenant-1"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "status", "failed"),
				),
			},
			{
				// The API never returns the secret; make sure that does not show up as drift.
				Config: testAccScalrGrafanaIntegrationLokiBearer(newName, testGrafanaFailingEndpoint),
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
				Config: testAccScalrGrafanaIntegrationLokiBearer(newName, testGrafanaOtherFailingEndpoint),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "endpoint", testGrafanaOtherFailingEndpoint),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "auth_type", "bearer"),
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "secret", "fake-token"),
				),
			},
			{
				// Back to no auth: tenant_id is cleared.
				Config: testAccScalrGrafanaIntegrationLokiNone(newName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_grafana_integration.test", "auth_type", "none"),
					resource.TestCheckNoResourceAttr("scalr_grafana_integration.test", "secret"),
					resource.TestCheckNoResourceAttr("scalr_grafana_integration.test", "tenant_id"),
				),
			},
		},
	})
}

func TestAccScalrGrafanaIntegrationResource_cloud(t *testing.T) {
	endpoint := os.Getenv("GRAFANA_ENDPOINT")
	instanceID := os.Getenv("GRAFANA_INSTANCE_ID")
	token := os.Getenv("GRAFANA_TOKEN")
	if endpoint == "" || instanceID == "" || token == "" {
		t.Skip("Grafana Cloud test skipped unless GRAFANA_ENDPOINT, GRAFANA_INSTANCE_ID and GRAFANA_TOKEN are set")
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
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "tempo"`, `auth_type = "none"`, testGrafanaFailingEndpoint, ""),
				ExpectError: regexp.MustCompile(`Attribute kind value must be one of`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "token"`, testGrafanaFailingEndpoint, ""),
				ExpectError: regexp.MustCompile(`Attribute auth_type value must be one of`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "none"`, "http://example.com/otlp", ""),
				ExpectError: regexp.MustCompile(`URL scheme must be https`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "none"`, "https://example.com/otlp/v1/logs/", ""),
				ExpectError: regexp.MustCompile(`Enter the base OTLP URL without /v1/logs`),
			},
			{
				Config: testAccScalrGrafanaIntegrationRaw(
					`kind = "cloud"`, `auth_type = "bearer"`, testGrafanaFailingEndpoint, `secret = "x"`,
				),
				ExpectError: regexp.MustCompile(`Grafana Cloud requires basic auth`),
			},
			{
				Config: testAccScalrGrafanaIntegrationRaw(
					`kind = "cloud"`, `auth_type = "basic"`, testGrafanaFailingEndpoint,
					"username = \"1\"\n  secret = \"x\"\n  tenant_id = \"t\"",
				),
				ExpectError: regexp.MustCompile(`"tenant_id" is only used with kind "loki"`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "basic"`, testGrafanaFailingEndpoint, `secret = "x"`),
				ExpectError: regexp.MustCompile(`"username" is required when "auth_type" is "basic"`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "bearer"`, testGrafanaFailingEndpoint, ""),
				ExpectError: regexp.MustCompile(`"secret" is required when "auth_type" is "bearer"`),
			},
			{
				Config:      testAccScalrGrafanaIntegrationRaw(`kind = "loki"`, `auth_type = "none"`, testGrafanaFailingEndpoint, `secret = "x"`),
				ExpectError: regexp.MustCompile(`"secret" is not used when "auth_type" is "none"`),
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

func testAccScalrGrafanaIntegrationLokiNone(name string) string {
	return fmt.Sprintf(`
resource "scalr_grafana_integration" "test" {
  name      = %q
  kind      = "loki"
  endpoint  = %q
  auth_type = "none"
}`, name, testGrafanaFailingEndpoint)
}

func testAccScalrGrafanaIntegrationLokiBearer(name, endpoint string) string {
	return fmt.Sprintf(`
resource "scalr_grafana_integration" "test" {
  name      = %q
  kind      = "loki"
  endpoint  = %q
  auth_type = "bearer"
  secret    = "fake-token"
  tenant_id = "tenant-1"
}`, name, endpoint)
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
