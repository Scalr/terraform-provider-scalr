package scalr

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scalr/go-scalr"
)

func TestAccCurrentRun_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccCurrentRunInitConfig(rInt),
			},
			{
				PreConfig: func() {
					_ = os.Unsetenv(currentRunIDEnvVar)
				},
				Config:   testAccCurrentRunDataSourceConfig(rInt),
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_current_run.test", "id", dummyIdentifier),
				),
			},
			{
				PreConfig: launchRun(fmt.Sprintf("test-env-%d", rInt), fmt.Sprintf("test-ws-%d", rInt)),
				Config:    testAccCurrentRunDataSourceConfig(rInt),
				PlanOnly:  true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.scalr_current_run.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_current_run.test", "workspace_name", fmt.Sprintf("test-ws-%d", rInt)),
				),
			},
		},
	})
}

func launchRun(environmentName, workspaceName string) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		options := GetEnvironmentByNameOptions{
			Name: &environmentName,
		}
		env, err := GetEnvironmentByName(ctx, options, scalrClient)
		if err != nil {
			log.Fatalf("Got error during environment fetching: %v", err)
			return
		}

		ws, err := scalrClient.Workspaces.Read(ctx, env.ID, workspaceName)
		if err != nil {
			log.Fatalf("Error retrieving workspace: %v", err)
		}

		cv, err := scalrClient.ConfigurationVersions.Create(ctx, scalr.ConfigurationVersionCreateOptions{
			Workspace: &scalr.Workspace{
				ID: ws.ID,
			},
		})

		if err != nil {
			log.Fatalf("Error creating cv: %v", cv)
		}

		run, err := scalrClient.Runs.Create(ctx, scalr.RunCreateOptions{
			Workspace: &scalr.Workspace{
				ID: ws.ID,
			},
			ConfigurationVersion: &scalr.ConfigurationVersion{
				ID: cv.ID,
			},
		})

		if err != nil {
			log.Fatalf("Error creating run: %v", err)
		}

		_ = os.Setenv(currentRunIDEnvVar, run.ID)
	}
}

func testAccCurrentRunInitConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name       = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}
`, rInt, defaultAccount)
}

func testAccCurrentRunDataSourceConfig(rInt int) string {
	return testAccCurrentRunInitConfig(rInt) + "data scalr_current_run test {}"
}
