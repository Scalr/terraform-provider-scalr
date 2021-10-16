package scalr

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	scalr "github.com/scalr/go-scalr"
)

func TestAccModuleVersionDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			//TODO:ape delete skip after SCALRCORE-19891
			t.Skip("Working on personal tocen but not working with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAccountModule(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_module.test", "id"),
				),
			},

			{
				PreConfig: waitForModuleVersions(fmt.Sprintf("test-env-%d", rInt)),
				Config:    testAccModuleVersionDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_module_version.latest", "id"),
					resource.TestCheckResourceAttrSet("data.scalr_module_version.latest", "version"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_module_version.latest", "source",
						"scalr_module.test", "source",
					),

					resource.TestCheckResourceAttrSet("data.scalr_module_version.version", "id"),
					resource.TestCheckResourceAttrSet("data.scalr_module_version.version", "version"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_module_version.version", "source",
						"scalr_module.test", "source",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_module_version.latest", "version",
						"data.scalr_module_version.version", "version",
					),
				),
			},
		},
	})
}

func waitForModuleVersions(environmentName string) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		env, err := GetEnvironmentByName(environmentName, scalrClient)
		if err != nil {
			log.Fatalf("Got error during environment fetching: %v", err)
			return
		}

		ml, err := scalrClient.Modules.List(ctx, scalr.ModuleListOptions{Environment: &env.ID})

		if len(ml.Items) == 0 {
			log.Fatalf("The test module for environment with name %s was not created: %v", environmentName, err)
		}
		var mID = ml.Items[0].ID

		for i := 0; i < 60; i++ {
			m, err := scalrClient.Modules.Read(ctx, mID)
			if err != nil {
				log.Fatalf("Error polling module  %s: %v", mID, err)
			}

			if m.Status != scalr.ModulePending {
				if m.Status != scalr.ModuleSetupComplete {
					log.Fatalf("Invalid module status %s", m.Status)
				}

				break
			}
			time.Sleep(time.Second)
		}
	}
}

func testAccScalrAccountModule(rInt int) string {
	return fmt.Sprintf(`
		locals {
			account_id = "%s"
		}
		resource scalr_vcs_provider test {
		  name       = "test-github-%[2]d"
		  vcs_type   = "%s"
		  token      = "%s"
		}
		
		resource scalr_environment test {
		  name       = "test-env-%[2]d"
		  account_id = local.account_id
		}
		
		resource "scalr_module" "test" {
		  environment_id = scalr_environment.test.id
		  account_id = local.account_id	
		  vcs_repo {
			identifier = "Scalr/terraform-scalr-revizor"
		  }
		  vcs_provider_id = scalr_vcs_provider.test.id
		}
`, defaultAccount, rInt, string(scalr.Github), GITHUB_TOKEN)
}

func testAccModuleVersionDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
		%s
		data "scalr_module_version" "latest" {
  			source = scalr_module.test.source
		}
		
		data "scalr_module_version" "version" {
  			source = scalr_module.test.source
			version = data.scalr_module_version.latest.version
		}
	`, testAccScalrAccountModule(rInt))
}
