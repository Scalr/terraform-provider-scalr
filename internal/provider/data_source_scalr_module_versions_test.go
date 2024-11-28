package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scalr/go-scalr"
)

func TestAccModuleVersionsDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testVcsAccGithubTokenPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_module_versions all_by_none {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,source` must be specified"),
				PlanOnly:    true,
			},
			{
				Config: testAccModuleVersionsReourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_module.test", "id"),
				),
			},
			{
				PreConfig: waitForModuleVersions(fmt.Sprintf("test-env-%d", rInt)),
				Config:    testAccModuleVersionsDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id", "id"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id", "source"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id", "versions.#", "2"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id", "versions.0.id"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id", "versions.0.status"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id", "versions.0.version"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id", "versions.0.status", "ok"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id", "versions.0.version", "0.0.2"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id", "versions.1.status", "ok"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id", "versions.1.version", "0.0.1"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_module_versions.all_by_id", "source",
						"scalr_module.test", "source",
					),

					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_source", "id"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_source", "source"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_source", "versions.#", "2"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_source", "versions.0.id"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_source", "versions.0.status"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_source", "versions.0.version"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_source", "versions.0.status", "ok"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_source", "versions.0.version", "0.0.2"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_source", "versions.1.status", "ok"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_source", "versions.1.version", "0.0.1"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_module_versions.all_by_source", "source",
						"scalr_module.test", "source",
					),

					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id_and_source", "id"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id_and_source", "source"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id_and_source", "versions.#", "2"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id_and_source", "versions.0.id"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id_and_source", "versions.0.status"),
					resource.TestCheckResourceAttrSet("data.scalr_module_versions.all_by_id_and_source", "versions.0.version"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id_and_source", "versions.0.status", "ok"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id_and_source", "versions.0.version", "0.0.2"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id_and_source", "versions.1.status", "ok"),
					resource.TestCheckResourceAttr("data.scalr_module_versions.all_by_id_and_source", "versions.1.version", "0.0.1"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_module_versions.all_by_id_and_source", "source",
						"scalr_module.test", "source",
					),
				),
			},
			{
				Config: testAccModuleVersionsDataSourceCustomConfig(rInt, `
				data scalr_module_versions bad_source {
				  id = scalr_module.test.id
				  source = "bad_source"
				}`),
				ExpectError: regexp.MustCompile("Could not find module with ID '.*?' and source 'bad_source'"),
				PlanOnly:    true,
			},
			{
				Config: testAccModuleVersionsDataSourceCustomConfig(rInt, `
				data scalr_module_versions bad_source {
				  id = "bad-id"
				  source = scalr_module.test.source
				}`),
				ExpectError: regexp.MustCompile("Could not find module with ID 'bad-id'"),
				PlanOnly:    true,
			},
		},
	})
}

func testAccModuleVersionsReourceConfig(rInt int) string {
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
`, defaultAccount, rInt, string(scalr.Github), githubToken)
}

func testAccModuleVersionsDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
		%s
		data "scalr_module_versions" "all_by_id" {
			id = scalr_module.test.id
		}

		data "scalr_module_versions" "all_by_source" {
			source = scalr_module.test.source
		}

		data "scalr_module_versions" "all_by_id_and_source" {
			id = scalr_module.test.id
			source = scalr_module.test.source
		}
	`, testAccModuleVersionsReourceConfig(rInt))
}

func testAccModuleVersionsDataSourceCustomConfig(rInt int, custom_config string) string {
	return fmt.Sprintf(`
		%s
		%s
	`, custom_config, testAccModuleVersionsReourceConfig(rInt))
}
