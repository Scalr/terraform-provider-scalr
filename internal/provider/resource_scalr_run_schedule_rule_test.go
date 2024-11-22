package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrRunScheduleRule_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrRunScheduleRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRunScheduleRuleBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrRunScheduleRuleExists("scalr_run_schedule_rule.test", &scalr.RunScheduleRule{}),
					resource.TestCheckResourceAttr(
						"scalr_run_schedule_rule.test", "schedule", "0 4 * * *"),
					resource.TestCheckResourceAttr(
						"scalr_run_schedule_rule.test", "schedule_mode", "apply"),
					resource.TestCheckResourceAttrSet(
						"scalr_run_schedule_rule.test", "workspace_id"),
				),
			},
		},
	})
}

func TestAccScalrRunScheduleRule_update(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrRunScheduleRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRunScheduleRuleBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrRunScheduleRuleExists("scalr_run_schedule_rule.test", &scalr.RunScheduleRule{}),
					resource.TestCheckResourceAttr(
						"scalr_run_schedule_rule.test", "schedule", "0 4 * * *"),
					resource.TestCheckResourceAttr(
						"scalr_run_schedule_rule.test", "schedule_mode", "apply"),
				),
			},
			{
				Config: testAccScalrRunScheduleRuleUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrRunScheduleRuleExists("scalr_run_schedule_rule.test", &scalr.RunScheduleRule{}),
					resource.TestCheckResourceAttr(
						"scalr_run_schedule_rule.test", "schedule", "0 5 * * *"),
					resource.TestCheckResourceAttr(
						"scalr_run_schedule_rule.test", "schedule_mode", "refresh"),
				),
			},
		},
	})
}

func TestAccScalrRunScheduleRule_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrRunScheduleRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrRunScheduleRuleBasic(rInt),
			},
			{
				ResourceName:      "scalr_run_schedule_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrRunScheduleRuleExists(resId string, rule *scalr.RunScheduleRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Run Schedule Rule ID is set")
		}

		scalrClient := testAccProvider.Meta().(*scalr.Client)
		r, err := scalrClient.RunScheduleRules.Read(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error reading run schedule rule: %v", err)
		}

		*rule = *r

		return nil
	}
}

func testAccCheckScalrRunScheduleRuleDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_run_schedule_rule" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Run Schedule Rule ID is set")
		}

		_, err := scalrClient.RunScheduleRules.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Run Schedule Rule %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrRunScheduleRuleBasic(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource "scalr_workspace" "test" {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource "scalr_run_schedule_rule" "test" {
  schedule      = "0 4 * * *"
  schedule_mode = "apply"
  workspace_id  = scalr_workspace.test.id
}`, rInt, defaultAccount)
}

func testAccScalrRunScheduleRuleUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource "scalr_workspace" "test" {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource "scalr_run_schedule_rule" "test" {
  schedule      = "0 5 * * *"
  schedule_mode = "refresh"
  workspace_id  = scalr_workspace.test.id
}`, rInt, defaultAccount)
}
