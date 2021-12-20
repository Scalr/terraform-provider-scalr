package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccScalrRunTriggersDataSource_basic(t *testing.T) {
	runTrigger := &scalr.RunTrigger{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRunTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRunTrigger_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRunTriggerExists("scalr_run_trigger.foobar", runTrigger),
					testAccCheckRunTriggerAttributes(runTrigger, "scalr_environment.test"),
				),
			},
		},
	})
}

func testAccCheckRunTriggerDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_run_trigger" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.RunTriggers.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("RunTrigger %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccRunTrigger_basic(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%d"
  account_id = "%s"
}

resource scalr_workspace upstream {
  name                  = "upstream-test"
  environment_id 		= scalr_environment.test.id
  auto_apply            = true
}

resource scalr_workspace downstream {
  name                  = "downstream-test"
  environment_id 		= scalr_environment.test.id
  auto_apply            = true
}

resource "scalr_run_trigger" "foobar" {
  downstream_id  = scalr_workspace.downstream.id
  upstream_id = scalr_workspace.upstream.id
}`, rInt, defaultAccount)
}

func testAccCheckRunTriggerExists(n string, runTrigger *scalr.RunTrigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		rt, err := scalrClient.RunTriggers.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*runTrigger = *rt

		return nil
	}
}

func testAccCheckRunTriggerAttributes(runTrigger *scalr.RunTrigger, environmentName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		environment, ok := s.RootModule().Resources[environmentName]
		if !ok {
			return fmt.Errorf("Not found: %s", environmentName)
		}

		downstreamID := runTrigger.Downstream.ID
		downstream, err := scalrClient.Workspaces.Read(ctx, environment.Primary.ID, "downstream-test")
		if err != nil {
			return fmt.Errorf("Error retreiving workspace downstream-test for environment %s, %v", environmentName, err)
		}

		if downstream.ID != downstreamID {
			return fmt.Errorf("Wrong downstream workspace ID: %v", downstream.ID)
		}

		upstreamID := runTrigger.Upstream.ID
		upstream, err := scalrClient.Workspaces.Read(ctx, environment.Primary.ID, "upstream-test")
		if err != nil {
			return fmt.Errorf("Error retreiving workspace upstream-test for environment %s, %v", environmentName, err)
		}
		if upstream.ID != upstreamID {
			return fmt.Errorf("Wrong upstream workspace ID: %v", upstream.ID)
		}

		return nil
	}
}
