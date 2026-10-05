package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"
)

// The tests use a fake API key, so the integration is expected to end up in the `failed` status.

func TestAccScalrDatadogIntegration_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-datadog")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrDatadogIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrDatadogIntegrationConfig(name, "fake-key", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_datadog_integration.test", "id"),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "name", name),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "api_key", "fake-key"),
					resource.TestCheckNoResourceAttr("scalr_datadog_integration.test", "deployment_url"),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "status", "failed"),
					resource.TestCheckResourceAttrSet("scalr_datadog_integration.test", "err_message"),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "account_id", defaultAccount),
				),
			},
			{
				// A server-set `failed` status does not cause drift.
				Config: testAccScalrDatadogIntegrationConfig(name, "fake-key", ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ResourceName:            "scalr_datadog_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func TestAccScalrDatadogIntegration_update(t *testing.T) {
	name := acctest.RandomWithPrefix("test-datadog")
	nameUpdated := acctest.RandomWithPrefix("test-datadog")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrDatadogIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrDatadogIntegrationConfig(name, "fake-key", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "name", name),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "status", "failed"),
				),
			},
			{
				Config: testAccScalrDatadogIntegrationConfig(nameUpdated, "fake-key-2", "https://api.datadoghq.eu"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "name", nameUpdated),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "api_key", "fake-key-2"),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "deployment_url", "https://api.datadoghq.eu"),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "status", "failed"),
				),
			},
			{
				Config: testAccScalrDatadogIntegrationConfig(nameUpdated, "fake-key-2", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("scalr_datadog_integration.test", "deployment_url"),
				),
			},
		},
	})
}

func TestAccScalrDatadogIntegration_importApiKey(t *testing.T) {
	if !isAccTest() {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)
	name := acctest.RandomWithPrefix("test-datadog")

	// Create the integration outside of Terraform, then adopt it with an import block.
	scalrClient := createScalrClientV2()
	di, err := scalrClient.DatadogIntegration.CreateDatadogIntegration(ctx, &schemas.DatadogIntegrationRequest{
		Attributes: schemas.DatadogIntegrationAttributesRequest{
			Name:   value.Set(name),
			ApiKey: value.Set("fake-key"),
		},
	})
	if err != nil {
		t.Fatalf("Error creating Datadog integration: %s", err)
	}
	t.Cleanup(func() {
		_ = scalrClient.DatadogIntegration.DeleteDatadogIntegration(ctx, di.ID)
	})

	config := fmt.Sprintf(`
import {
  to = scalr_datadog_integration.test
  id = %q
}
`, di.ID) + testAccScalrDatadogIntegrationConfig(name, "fake-key", "")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrDatadogIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				// The API key is not returned by the API, so the first apply after import sets it...
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("scalr_datadog_integration.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "id", di.ID),
					resource.TestCheckResourceAttr("scalr_datadog_integration.test", "api_key", "fake-key"),
				),
			},
			{
				// ...and there is no drift afterwards.
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccScalrDatadogIntegrationConfig(name, apiKey, deploymentURL string) string {
	url := ""
	if deploymentURL != "" {
		url = fmt.Sprintf("deployment_url = %q", deploymentURL)
	}
	return fmt.Sprintf(`
resource "scalr_datadog_integration" "test" {
  name    = %q
  api_key = %q
  %s
}`, name, apiKey, url)
}

func testAccCheckScalrDatadogIntegrationDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_datadog_integration" {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}
		_, err := scalrClient.DatadogIntegration.GetDatadogIntegration(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Datadog integration %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, client.ErrNotFound) {
			return err
		}
	}

	return nil
}
