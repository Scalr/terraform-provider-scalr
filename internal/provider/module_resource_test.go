package provider

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
	"github.com/scalr/go-scalr/v2/scalr/client"
)

func TestAccScalrModule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			//TODO:ape delete skip after SCALRCORE-19891
			t.Skip("Working on personal token but not working with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModulesOnAllScopes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleExists("scalr_module.test"),
					resource.TestCheckResourceAttr("scalr_module.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttrSet("scalr_module.test", "environment_id"),
					resource.TestCheckResourceAttr("scalr_module.test", "vcs_repo.0.identifier", "Scalr/terraform-scalr-revizor"),

					testAccCheckScalrModuleExists("scalr_module.test-account"),
					resource.TestCheckResourceAttr("scalr_module.test-account", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_module.test-account", "vcs_repo.0.identifier", "Scalr/terraform-scalr-revizor"),

					testAccCheckScalrModuleExists("scalr_module.test-global"),
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
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleDestroy,
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

func TestAccScalrModule_withNamespace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleWithNamespace(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleExists("scalr_module.test"),
					resource.TestCheckResourceAttrSet("scalr_module.test", "namespace_id"),
				),
			},
		},
	})
}

func TestAccScalrModule_UpgradeFromSDK(t *testing.T) {
	config := testAccScalrModuleWithNamespace()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testVcsAccGithubTokenPreCheck(t)
		},
		CheckDestroy: testAccCheckScalrModuleDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"scalr": {
						Source:            "registry.scalr.io/scalr/scalr",
						VersionConstraint: "3.19.0",
					},
				},
				Config: config,
				Check:  resource.TestCheckResourceAttrSet("scalr_module.test", "id"),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(t),
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccScalrModule_oci(t *testing.T) {
	registryURL, username, password := dockerRegistryTestCreds(t)
	rInd := GetRandomInteger()
	config := testAccScalrDockerIntegrationConfig(fmt.Sprintf("test-docker-%d", rInd), registryURL, username, password, false) +
		fmt.Sprintf(`
resource scalr_module_namespace test {
  name = "test-namespace-%d"
}

resource "scalr_module" "test" {
  namespace_id          = scalr_module_namespace.test.id
  docker_integration_id = scalr_docker_integration.test.id
  docker_image          = "%s/terraform-null-wait"
  name                  = "wait"
  module_provider       = "null"
}`, rInd, username)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleExists("scalr_module.test"),
					resource.TestCheckResourceAttr("scalr_module.test", "source_type", "docker"),
					resource.TestCheckResourceAttr("scalr_module.test", "docker_image", username+"/terraform-null-wait"),
					resource.TestCheckResourceAttrPair("scalr_module.test", "docker_integration_id", "scalr_docker_integration.test", "id"),
					resource.TestCheckResourceAttr("scalr_module.test", "name", "wait"),
					resource.TestCheckResourceAttr("scalr_module.test", "module_provider", "null"),
					resource.TestCheckNoResourceAttr("scalr_module.test", "vcs_provider_id"),
					resource.TestCheckResourceAttr("scalr_module.test", "vcs_repo.#", "0"),
				),
			},
			{
				ResourceName:      "scalr_module.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The module status changes asynchronously while Scalr syncs the module versions.
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
	})
}

func TestAccScalrModule_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scalr_module" "test" {
				  name                  = "network"
				  module_provider       = "aws"
				  docker_image          = "scalr/terraform-aws-network"
				  docker_integration_id = "docker-xxxxx"
				  vcs_provider_id       = "vcs-xxxxx"
				  vcs_repo {
					identifier = "Scalr/terraform-scalr-revizor"
				  }
				}
				`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
			{
				Config: `
				resource "scalr_module" "test" {
				  module_provider = "aws"
				}
				`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
			{
				Config: `
				resource "scalr_module" "test" {
				  vcs_repo {
					identifier = "Scalr/terraform-scalr-revizor"
				  }
				}
				`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
			{
				Config: `
				resource "scalr_module" "test" {
				  docker_image          = "scalr/terraform-aws-network"
				  docker_integration_id = "docker-xxxxx"
				}
				`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("is required for OCI-sourced modules"),
			},
			{
				Config: `
				resource "scalr_module" "test" {
				  name                  = "network"
				  module_provider       = "aws"
				  docker_image          = "scalr/terraform-aws-network"
				}
				`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

func testAccCheckScalrModuleExists(resId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Module.GetModule(ctx, rs.Primary.ID, nil)
		return err
	}
}

func testAccCheckScalrModuleDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_module" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Module.GetModule(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Module %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, client.ErrNotFound) {
			return err
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

func testAccScalrModuleWithNamespace() string {
	rInd := GetRandomInteger()

	return fmt.Sprintf(`
		resource scalr_vcs_provider test {
		  name       = "test-github-provider-namespace-%[1]d"
		  vcs_type   = "%s"
		  token      = "%s"
		}
		
		resource scalr_module_namespace test {
		  name = "test-namespace-%[1]d"
		}
		
		resource "scalr_module" "test" {
		  namespace_id = scalr_module_namespace.test.id
		  vcs_repo {
			identifier = "Scalr/terraform-scalr-revizor"
		  }
		  vcs_provider_id = scalr_vcs_provider.test.id
		}
`, rInd, string(scalr.Github), githubToken)
}
