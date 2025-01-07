package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrWorkspace_basic(t *testing.T) {
	workspace := &scalr.Workspace{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "deletion_protection_enabled", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "type", "unmapped"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "auto_queue_runs", string(scalr.AutoQueueRunsModeAlways)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "execution_mode", string(scalr.WorkspaceExecutionModeRemote)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", ""),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "run_operation_timeout", "18"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "var_files.0", "test1.tfvars"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "var_files.1", "test2.tfvars"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_init", "./scripts/pre-init.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_plan", "./scripts/pre-plan.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.post_plan", "./scripts/post-plan.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_apply", "./scripts/pre-apply.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.post_apply", "./scripts/post-apply.sh"),
					resource.TestCheckResourceAttrSet("scalr_workspace.test", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("scalr_workspace.test", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("scalr_workspace.test", "created_by.0.username"),
				),
			},
		},
	})
}

func TestAccScalrWorkspace_create_missed_vcs_attr(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrWorkspaceMissedVcsProvider(rInt),
				ExpectError: regexp.MustCompile("\"vcs_repo\": all of `vcs_provider_id,vcs_repo` must be specified"),
			},
			{
				Config:      testAccScalrWorkspaceMissedVcsRepo(rInt),
				ExpectError: regexp.MustCompile("\"vcs_provider_id\": all of `vcs_provider_id,vcs_repo` must be specified"),
			},
		},
	})
}

func TestAccScalrWorkspace_monorepo(t *testing.T) {
	workspace := &scalr.Workspace{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceMonorepo(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceMonorepoAttributes(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "name", "workspace-monorepo"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "auto_queue_runs", string(scalr.AutoQueueRunsModeNever)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "execution_mode", string(scalr.WorkspaceExecutionModeRemote)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", "/db"),
					resource.TestCheckNoResourceAttr("scalr_workspace.test", "run_operation_timeout"),
				),
			},
		},
	})
}

func TestAccScalrWorkspace_update(t *testing.T) {
	workspace := &scalr.Workspace{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists("scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr("scalr_workspace.test", "name", "workspace-test"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "auto_apply", "true"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "deletion_protection_enabled", "false"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "execution_mode", string(scalr.WorkspaceExecutionModeRemote)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "auto_queue_runs", string(scalr.AutoQueueRunsModeAlways)),
					resource.TestCheckResourceAttr("scalr_workspace.test", "working_directory", ""),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "run_operation_timeout", "18"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "var_files.0", "test1.tfvars"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "var_files.1", "test2.tfvars"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_init", "./scripts/pre-init.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_plan", "./scripts/pre-plan.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.post_plan", "./scripts/post-plan.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_apply", "./scripts/pre-apply.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.post_apply", "./scripts/post-apply.sh"),
				),
			},

			{
				Config: testAccScalrWorkspaceUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceAttributesUpdated(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "deletion_protection_enabled", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "operations", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "execution_mode", string(scalr.WorkspaceExecutionModeLocal)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "terraform_version", "1.1.9"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "iac_platform", string(scalr.WorkspaceIaCPlatformTerraform)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", "terraform/test"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "run_operation_timeout", "200"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "type", "staging"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "var_files.0", "test1updated.tfvars"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "var_files.1", "test2updated.tfvars"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_init", "./scripts/pre-init_updated.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_plan", "./scripts/pre-plan_updated.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.post_plan", "./scripts/post-plan_updated.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.pre_apply", "./scripts/pre-apply_updated.sh"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "hooks.0.post_apply", "./scripts/post-apply_updated.sh"),
				),
			},

			{
				Config: testAccScalrWorkspaceUpdateWithoutHooks(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceAttributesUpdated(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "operations", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "execution_mode", string(scalr.WorkspaceExecutionModeLocal)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "terraform_version", "1.1.9"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "iac_platform", string(scalr.WorkspaceIaCPlatformTerraform)),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", "terraform/test"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "run_operation_timeout", "0"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "hooks.#", "0"),
				),
			},

			{
				Config: testAccScalrWorkspaceUpdateWorkingDir(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.test", workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", ""),
				),
			},
		},
	})
}

func TestAccScalrWorkspace_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceBasic(rInt),
			},

			{
				ResourceName:      "scalr_workspace.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalrWorkspace_providerConfiguration(t *testing.T) {
	workspace := &scalr.Workspace{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceProviderConfiguration(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceProviderConfigurations(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "provider_configuration.#", "3"),
				),
			},
			{
				Config: testAccScalrWorkspaceProviderConfigurationUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceProviderConfigurationsUpdated(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "provider_configuration.#", "3"),
				),
			},
		},
	})
}

func TestAccScalrWorkspaceSSHKey(t *testing.T) {
	workspace := &scalr.Workspace{}
	sshKey := &scalr.SSHKey{}
	rInt := GetRandomInteger()
	sshKeyName := fmt.Sprintf("ssh-key-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceWithSSHKeyConfig(rInt, sshKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists("scalr_workspace.test", workspace),
					testAccCheckScalrSSHKeyExists("scalr_ssh_key.test", sshKey),
					resource.TestCheckResourceAttr("scalr_workspace.test", "name", "workspace-with-ssh-key"),
					resource.TestCheckResourceAttr("scalr_ssh_key.test", "name", sshKeyName),
					resource.TestCheckResourceAttrSet("scalr_workspace.test", "ssh_key_id"),
				),
			},
		},
	})
}

func testAccCheckScalrSSHKeyExists(n string, sshKey *scalr.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSH key ID is set")
		}

		k, err := scalrClient.SSHKeys.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*sshKey = *k
		return nil
	}
}

func TestAccScalrWorkspace_emptyHooks(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceEmptyHooks(rInt),
			},
		},
	})
}

func testAccCheckScalrWorkspaceExists(
	n string, workspace *scalr.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the workspace
		w, err := scalrClient.Workspaces.ReadByID(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*workspace = *w

		return nil
	}
}

func testAccCheckScalrWorkspaceAttributes(
	workspace *scalr.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-test" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.AutoApply != true {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.ExecutionMode != scalr.WorkspaceExecutionModeRemote {
			return fmt.Errorf("Bad execution mode: %s", workspace.ExecutionMode)
		}

		if workspace.WorkingDirectory != "" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckScalrWorkspaceMonorepoAttributes(
	workspace *scalr.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-monorepo" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.WorkingDirectory != "/db" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckScalrWorkspaceAttributesUpdated(
	workspace *scalr.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-updated" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.AutoApply != false {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.ExecutionMode != scalr.WorkspaceExecutionModeLocal {
			return fmt.Errorf("Bad execution mode: %s", workspace.ExecutionMode)
		}

		if workspace.TerraformVersion != "1.1.9" {
			return fmt.Errorf("Bad Terraform version: %s", workspace.TerraformVersion)
		}

		if workspace.WorkingDirectory != "terraform/test" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckScalrWorkspaceProviderConfigurations(
	workspace *scalr.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		links, err := getProviderConfigurationWorkspaceLinks(ctx, scalrClient, workspace.ID)
		if err != nil {
			return fmt.Errorf("Error retrieving provider configuration links: %v", err)
		}

		if len(links) != 3 {
			return fmt.Errorf("Bad provider configurations: %v", links)
		}

		pcfgNameToAliases := make(map[string][]string)
		for _, currentLink := range links {
			pcfgNameToAliases[currentLink.ProviderConfiguration.Name] = append(
				pcfgNameToAliases[currentLink.ProviderConfiguration.Name], currentLink.Alias,
			)
		}
		if aliases, ok := pcfgNameToAliases["kubernetes"]; ok {
			if !(len(aliases) == 1 && aliases[0] == "") {
				return fmt.Errorf("Bad kubernetes link aliases: %v", aliases)
			}
		} else {
			return fmt.Errorf("Kubernetes provider configuration link doesn't exist.")
		}

		if aliases, ok := pcfgNameToAliases["consul"]; ok {
			if len(aliases) != 2 {
				return fmt.Errorf("Bad consul provider configuration link aliases: %v", aliases)
			}
			expected := []string{"dev", ""}
			for _, expectedAlias := range expected {
				found := false
				for _, gotAlias := range aliases {
					if expectedAlias == gotAlias {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("Bad consul provider configuration link aliases: %v", aliases)
				}
			}

		} else {
			return fmt.Errorf("Bad consul provider configuration link aliases: %v", aliases)
		}

		return nil
	}
}

func testAccCheckScalrWorkspaceProviderConfigurationsUpdated(
	workspace *scalr.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		links, err := getProviderConfigurationWorkspaceLinks(ctx, scalrClient, workspace.ID)
		if err != nil {
			return fmt.Errorf("Error retrieving provider configuration links: %v", err)
		}

		if len(links) != 3 {
			return fmt.Errorf("Bad provider configurations: %v", links)
		}

		pcfgNameToAliases := make(map[string][]string)
		for _, currentLink := range links {
			pcfgNameToAliases[currentLink.ProviderConfiguration.Name] = append(
				pcfgNameToAliases[currentLink.ProviderConfiguration.Name], currentLink.Alias,
			)
		}
		if aliases, ok := pcfgNameToAliases["kubernetes"]; ok {
			if !(len(aliases) == 1 && aliases[0] == "") {
				return fmt.Errorf("Bad kubernetes link aliases: %v", aliases)
			}
		} else {
			return fmt.Errorf("Kubernetes provider configuration link doesn't exist.")
		}

		if aliases, ok := pcfgNameToAliases["consul"]; ok {
			if len(aliases) != 2 {
				return fmt.Errorf("Bad consul provider configuration link aliases: %v", aliases)
			}

			expected := []string{"dev", "dev2"}
			for _, expectedAlias := range expected {
				found := false
				for _, gotAlias := range aliases {
					if expectedAlias == gotAlias {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("Bad consul provider configuration link aliases: %v", aliases)
				}
			}

		} else {
			return fmt.Errorf("Bad consul provider configuration link aliases: %v", aliases)
		}

		return nil
	}
}

func testAccCheckScalrWorkspaceDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_workspace" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Workspaces.ReadByID(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Workspace %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccScalrWorkspaceCommonConfig = `
resource scalr_environment test {
  name       = "test-env-%d"
  account_id = "%s"
}
%s
`

func testAccScalrWorkspaceEmptyHooks(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount, `
resource scalr_workspace test {
  name                   = "workspace-test"
  environment_id         = scalr_environment.test.id
  hooks {
    pre_init   = ""
    pre_plan   = ""
    post_plan  = ""
    pre_apply  = ""
    post_apply = ""
  }
}`)
}

func testAccScalrWorkspaceBasic(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount, `
resource scalr_workspace test {
  name                           = "workspace-test"
  environment_id                 = scalr_environment.test.id
  auto_apply                     = true
  run_operation_timeout          = 18
  var_files                      = ["test1.tfvars", "test2.tfvars"]
  auto_queue_runs                = "always"
  deletion_protection_enabled    = false
  type                           = "unmapped"
  hooks {
    pre_init   = "./scripts/pre-init.sh"
    pre_plan   = "./scripts/pre-plan.sh"
    post_plan  = "./scripts/post-plan.sh"
    pre_apply  = "./scripts/pre-apply.sh"
    post_apply = "./scripts/post-apply.sh"
  }
}`)
}

func testAccScalrWorkspaceMonorepo(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount, `
resource "scalr_workspace" "test" {
  name                  = "workspace-monorepo"
  environment_id 		= scalr_environment.test.id
  working_directory     = "/db"
  auto_queue_runs       = "never"
}`)
}

func testAccScalrWorkspaceUpdate(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount,
		fmt.Sprintf(`
resource "scalr_workspace" "test" {
  name                          = "workspace-updated"
  environment_id 		        = scalr_environment.test.id
  auto_apply                    = false
  execution_mode                = "%s"
  terraform_version             = "1.1.9"
  working_directory             = "terraform/test"
  run_operation_timeout         = 200
  deletion_protection_enabled   = true
  var_files                     = ["test1updated.tfvars", "test2updated.tfvars"]
  type                          = "staging"
  hooks {
    pre_init   = "./scripts/pre-init_updated.sh"
    pre_plan   = "./scripts/pre-plan_updated.sh"
    post_plan  = "./scripts/post-plan_updated.sh"
    pre_apply  = "./scripts/pre-apply_updated.sh"
    post_apply = "./scripts/post-apply_updated.sh"
  }
}`, scalr.WorkspaceExecutionModeLocal),
	)
}

func testAccScalrWorkspaceMissedVcsProvider(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount,
		fmt.Sprintf(`
resource "scalr_workspace" "test" {
  name                  = "workspace-updated"
  environment_id 		= scalr_environment.test.id
  auto_apply            = false
  execution_mode        = "%s"
  terraform_version     = "1.1.9"
  working_directory     = "terraform/test"
  vcs_repo {
   identifier = "TestRepo/local"
   branch     = "main"
  }
}`, scalr.WorkspaceExecutionModeLocal),
	)
}

func testAccScalrWorkspaceMissedVcsRepo(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount,
		fmt.Sprintf(`
resource "scalr_workspace" "test" {
  name                  = "workspace-updated"
  environment_id 		= scalr_environment.test.id
  auto_apply            = false
  execution_mode        = "%s"
  terraform_version     = "1.1.9"
  working_directory     = "terraform/test"
  vcs_provider_id	    = "test_provider_id"
}`, scalr.WorkspaceExecutionModeLocal),
	)
}

func testAccScalrWorkspaceUpdateWithoutHooks(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount,
		fmt.Sprintf(`
resource "scalr_workspace" "test" {
  name                  = "workspace-updated"
  environment_id 		= scalr_environment.test.id
  auto_apply            = false
  execution_mode        = "%s"
  terraform_version     = "1.1.9"
  working_directory     = "terraform/test"
}`, scalr.WorkspaceExecutionModeLocal),
	)
}

func testAccScalrWorkspaceUpdateWorkingDir(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount,
		fmt.Sprintf(`
resource "scalr_workspace" "test" {
  name                  = "workspace-updated"
  environment_id 		= scalr_environment.test.id
  auto_apply            = false
  execution_mode        = "%s"
  terraform_version     = "1.1.9"
  working_directory     = ""
  hooks {
    pre_init   = "./scripts/pre-init_updated.sh"
    pre_plan   = "./scripts/pre-plan_updated.sh"
    post_plan  = "./scripts/post-plan_updated.sh"
    pre_apply  = "./scripts/pre-apply_updated.sh"
    post_apply = "./scripts/post-apply_updated.sh"
  }
}`, scalr.WorkspaceExecutionModeLocal),
	)
}

func testAccScalrWorkspaceProviderConfiguration(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount,
		fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name         = "kubernetes"
  account_id   = scalr_environment.test.account_id
  environments = ["*"]
  custom {
    provider_name = "kubernetes"
    argument {
      name  = "config_path"
      value = "~/.kube/config"
	}
  }
}

resource "scalr_provider_configuration" "consul" {
  name         = "consul"
  account_id   = scalr_environment.test.account_id
  environments = ["*"]
  custom {
    provider_name = "consul"
    argument {
      name  = "config_path"
      value = "~/.kube/config"
    }
  }
}

resource "scalr_workspace" "test" {
  name                   = "workspace-pcfg-test"
  environment_id         = scalr_environment.test.id
  auto_apply             = false
  execution_mode        = "%s"
  working_directory      = "terraform/test"
  provider_configuration {
    id = scalr_provider_configuration.kubernetes.id
  }
  provider_configuration {
    id    = scalr_provider_configuration.consul.id
    alias = "dev"
  }
  provider_configuration {
    id    = scalr_provider_configuration.consul.id
    alias = ""
  }
}`, scalr.WorkspaceExecutionModeLocal),
	)
}

func testAccScalrWorkspaceProviderConfigurationUpdated(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount,
		fmt.Sprintf(`
resource "scalr_provider_configuration" "kubernetes" {
  name         = "kubernetes"
  account_id   = scalr_environment.test.account_id
  environments = ["*"]
  custom {
    provider_name = "kubernetes"
    argument {
      name  = "config_path"
      value = "~/.kube/config"
    }
  }
}

resource "scalr_provider_configuration" "consul" {
  name         = "consul"
  account_id   = scalr_environment.test.account_id
  environments = ["*"]
  custom {
    provider_name = "consul"
    argument {
      name  = "config_path"
      value = "~/.kube/config"
	}
  }
}

resource "scalr_workspace" "test" {
  name                   = "workspace-pcfg-test"
  environment_id         = scalr_environment.test.id
  auto_apply             = false
  execution_mode        = "%s"
  working_directory      = "terraform/test"
  provider_configuration {
    id    = scalr_provider_configuration.kubernetes.id
    alias = ""
  }
  provider_configuration {
    id    = scalr_provider_configuration.consul.id
    alias = "dev"
  }
  provider_configuration {
    id    = scalr_provider_configuration.consul.id
    alias = "dev2"
  }
}`, scalr.WorkspaceExecutionModeLocal),
	)
}

func testAccScalrWorkspaceWithSSHKeyConfig(rInt int, sshKeyName string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "%s"
}

resource "scalr_ssh_key" "test" {
  name        = "%s"
  private_key = <<EOF
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBvMDyNaYtWK2TmJIfFhmPZeGxK0bWnNDhjlTZ+V6e4x
-----END PRIVATE KEY-----
EOF
  account_id  = scalr_environment.test.account_id
  environments = [scalr_environment.test.id]
}

resource "scalr_workspace" "test" {
  name                   = "workspace-with-ssh-key"
  environment_id         = scalr_environment.test.id
  ssh_key_id             = scalr_ssh_key.test.id
  auto_apply             = true
  execution_mode         = "remote"
  working_directory      = ""
}`, rInt, defaultAccount, sshKeyName)
}
