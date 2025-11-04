package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestDriftDetection_basic(t *testing.T) {
	envName := acctest.RandomWithPrefix("test-env")
	resourceName := "scalr_drift_detection.test"
	var driftDetectionID string

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testDriftDetectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      `resource "scalr_drift_detection" "test" {}`,
				ExpectError: regexp.MustCompile("The argument \"check_period\" is required, but no definition was found"),
				PlanOnly:    true,
			},
			{
				Config:      `resource "scalr_drift_detection" "test" {}`,
				ExpectError: regexp.MustCompile("The argument \"environment_id\" is required, but no definition was found"),
				PlanOnly:    true,
			},
			{
				Config:      testDriftDetectionConfig(envName, "bad"),
				ExpectError: regexp.MustCompile(`Attribute check_period value must be one of: \["daily" "weekly"], got: "bad"`),
				PlanOnly:    true,
			},
			{
				Config: testDriftDetectionConfig(envName, "daily"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttr(resourceName, "check_period", "daily"),
					testDriftDetectionSaveID(resourceName, &driftDetectionID),
				),
			},
			{
				Config: testDriftDetectionConfig(envName, "weekly"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttr(resourceName, "check_period", "weekly"),
				),
			},
			{
				Config: testDriftDetectionDeleteConfig(envName),
				Check: resource.ComposeTestCheckFunc(
					testDriftDetectionDeleted(resourceName, &driftDetectionID),
				),
			},
		},
	})
}

func TestDriftDetection_import(t *testing.T) {
	envName := acctest.RandomWithPrefix("test-env")
	resourceName := "scalr_drift_detection.test"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testDriftDetectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDriftDetectionConfig(envName, "daily"),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testDriftDetectionDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_drift_detection" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.DriftDetections.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("DriftDetection %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testDriftDetectionDeleted(name string, driftDetectionID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		_, ok := s.RootModule().Resources[name]
		if ok {
			return fmt.Errorf("DriftDection resource %s still exist in the state", name)
		}

		_, err := scalrClient.DriftDetections.Read(ctx, *driftDetectionID)
		if err == nil {
			return fmt.Errorf("DriftDetection %s still exists", *driftDetectionID)
		}

		return nil
	}
}

func testDriftDetectionSaveID(name string, driftDetectionID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("DriftDetection resource found: %s", name)
		}
		*driftDetectionID = rs.Primary.ID

		return nil
	}
}

func testDriftDetectionConfig(envName string, checkPeriod string) string {
	return fmt.Sprintf(testDriftDetectionConfigBase, envName, defaultAccount, fmt.Sprintf(`
resource "scalr_drift_detection" "test" {
  environment_id = scalr_environment.test.id
  check_period = "%s"
}
`, checkPeriod))
}

func testDriftDetectionDeleteConfig(envName string) string {
	return fmt.Sprintf(testDriftDetectionConfigBase, envName, defaultAccount, "")
}

const testDriftDetectionConfigBase = `
resource "scalr_environment" "test" {
  name       = "%s"
  account_id = "%s"
}
%s
`
