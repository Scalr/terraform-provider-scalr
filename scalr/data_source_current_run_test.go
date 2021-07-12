package scalr

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	scalr "github.com/scalr/go-scalr"
)

func TestAccCurrentRun_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					os.Unsetenv(currentRunIDEnvVar)
				},
				Config: testAccCurrentRunDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.scalr_current_run.test", "id"),
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
		var environmentID *string
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		envl, err := scalrClient.Environments.List(ctx)
		if err != nil {
			log.Fatalf("Error retrieving environments: %v", err)
		}

		for _, env := range envl.Items {
			if env.Name == environmentName {
				environmentID = &env.ID
				break
			}
		}
		if environmentID == nil {
			log.Fatalf("Could not find environment with name: %s", environmentName)
			return
		}

		ws, err := scalrClient.Workspaces.Read(ctx, *environmentID, workspaceName)
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

		os.Setenv(currentRunIDEnvVar, run.ID)
	}
}

func testAccCurrentRunDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name       = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

data scalr_current_run test {
}`, rInt, defaultAccount)
}
