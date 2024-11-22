package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrModule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			//TODO:ape delete skip after SCALRCORE-19891
			t.Skip("Working on personal token but not working with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModulesOnAllScopes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleExists("scalr_module.test", &scalr.Module{}),
					resource.TestCheckResourceAttr("scalr_module.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttrSet("scalr_module.test", "environment_id"),
					resource.TestCheckResourceAttr("scalr_module.test", "vcs_repo.0.identifier", "Scalr/terraform-scalr-revizor"),

					testAccCheckScalrModuleExists("scalr_module.test-account", &scalr.Module{}),
					resource.TestCheckResourceAttr("scalr_module.test-account", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_module.test-account", "vcs_repo.0.identifier", "Scalr/terraform-scalr-revizor"),

					testAccCheckScalrModuleExists("scalr_module.test-global", &scalr.Module{}),
					resource.TestCheckResourceAttr("scalr_module.test-global", "vcs_repo.0.identifier", "Scalr/terraform-scalr-revizor"),
				),
			},
			{
				Config: `
				resource "scalr_module" "test-not-valid" {
				  vcs_repo {
					identifier = "Scalr/terraform-scalr-revizor"
				  }
				  vcs_provider_id = "vcs-xxxxx"
				}
				`,
				ExpectError: regexp.MustCompile("VcsProvider with ID 'vcs-xxxxx' not found or user unauthorized"),
			},
			{
				Config: `
				resource "scalr_module" "test-not-valid" {
				  vcs_repo {
					identifier = "Scalr/terraform-scalr-revizor"
				  }
				  vcs_provider_id = "vcs-xxxxx"
				  environment_id ="env-test"	
				}
				`,
				ExpectError: regexp.MustCompile("The attribute account_id is required"),
			},
		},
	})
}

func TestAccScalrModule_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testVcsAccGithubTokenPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModule(),
			},
			{
				ResourceName:      "scalr_module.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrModuleExists(moduleId string, module *scalr.Module) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[moduleId]
		if !ok {
			return fmt.Errorf("Not found: %s", moduleId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the module
		m, err := scalrClient.Modules.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*module = *m

		return nil
	}
}

func testAccCheckScalrModuleDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_module" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Modules.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Module %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrModule() string {
	return fmt.Sprintf(`
	resource scalr_vcs_provider test {
	  name       = "test-github-provider-import"
	  vcs_type   = "%s"
	  token      = "%s"
	}
	
	resource "scalr_module" "test" {
	  vcs_repo {
		identifier = "Scalr/terraform-scalr-revizor"
	  }
	  vcs_provider_id = scalr_vcs_provider.test.id
}
`, string(scalr.Github), githubToken)
}

func testAccScalrModulesOnAllScopes() string {
	rInd := GetRandomInteger()

	return fmt.Sprintf(`
		resource scalr_vcs_provider test {
		  name       = "test-github-provider-all-scopes-%[1]d"
		  vcs_type   = "%s"
		  token      = "%s"
		}
		
		locals {
			account_id = "%s"
		}
		
		resource "scalr_module" "test-global" {
		  vcs_repo {
			identifier = "Scalr/terraform-scalr-revizor"
		  }
		  vcs_provider_id = scalr_vcs_provider.test.id
		}
		
		resource "scalr_module" "test-account" {
		  account_id = local.account_id
		  vcs_repo {
			identifier = "Scalr/terraform-scalr-revizor"
		  }
		  vcs_provider_id = scalr_vcs_provider.test.id
		}
		
		resource scalr_environment test {
		  name       = "test-env-for-module-%[1]d"
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
`, rInd, string(scalr.Github), githubToken, defaultAccount)
}
