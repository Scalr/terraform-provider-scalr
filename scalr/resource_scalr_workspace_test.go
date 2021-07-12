package scalr

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccScalrWorkspace_basic(t *testing.T) {
	workspace := &scalr.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
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
						"scalr_workspace.test", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", ""),
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

func TestAccScalrWorkspace_monorepo(t *testing.T) {
	workspace := &scalr.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
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
						"scalr_workspace.test", "working_directory", "/db"),
				),
			},
		},
	})
}

func TestAccScalrWorkspace_renamed(t *testing.T) {
	workspace := &scalr.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
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
						"scalr_workspace.test", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", ""),
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
				PreConfig: testAccCheckScalrWorkspaceRename(fmt.Sprintf("test-env-%d", rInt), "workspace-test"),
				Config:    testAccScalrWorkspaceRenamed(rInt),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", ""),
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
		},
	})
}
func TestAccScalrWorkspace_update(t *testing.T) {
	workspace := &scalr.Workspace{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists("scalr_workspace.test", workspace),
					testAccCheckScalrWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr("scalr_workspace.test", "name", "workspace-test"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "auto_apply", "true"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "operations", "true"),
					resource.TestCheckResourceAttr("scalr_workspace.test", "working_directory", ""),
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
						"scalr_workspace.test", "operations", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "terraform_version", "0.12.19"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.test", "working_directory", "terraform/test"),
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
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
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

func testAccCheckScalrWorkspaceExists(
	n string, workspace *scalr.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

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

		if workspace.Operations != true {
			return fmt.Errorf("Bad operations: %t", workspace.Operations)
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

func testAccCheckScalrWorkspaceRename(environmentName, workspaceName string) func() {
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

		w, err := scalrClient.Workspaces.Update(
			context.Background(),
			ws.ID,
			scalr.WorkspaceUpdateOptions{Name: scalr.String("renamed-out-of-band")},
		)
		if err != nil {
			log.Fatalf("Could not rename the workspace out of band: %v", err)
		}

		if w.Name != "renamed-out-of-band" {
			log.Fatalf("Failed to rename the workspace out of band: %v", err)
		}
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

		if workspace.Operations != false {
			return fmt.Errorf("Bad operations: %t", workspace.Operations)
		}

		if workspace.TerraformVersion != "0.12.19" {
			return fmt.Errorf("Bad Terraform version: %s", workspace.TerraformVersion)
		}

		if workspace.WorkingDirectory != "terraform/test" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckScalrWorkspaceDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

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

func testAccScalrWorkspaceBasic(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount, `
resource scalr_workspace test {
  name           = "workspace-test"
  environment_id = scalr_environment.test.id
  auto_apply     = true
  hooks {
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
}`)
}

func testAccScalrWorkspaceRenamed(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount, `
resource "scalr_workspace" "test" {
  name           = "renamed-out-of-band"
  environment_id = scalr_environment.test.id
  auto_apply     = true
  hooks {
    pre_plan   = "./scripts/pre-plan.sh"
    post_plan  = "./scripts/post-plan.sh"
    pre_apply  = "./scripts/pre-apply.sh"
    post_apply = "./scripts/post-apply.sh"
  }
}`)
}

func testAccScalrWorkspaceUpdate(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount, `
resource "scalr_workspace" "test" {
  name                  = "workspace-updated"
  environment_id 		= scalr_environment.test.id
  auto_apply            = false
  operations            = false
  terraform_version     = "0.12.19"
  working_directory     = "terraform/test"
  hooks {
    pre_plan   = "./scripts/pre-plan_updated.sh"
    post_plan  = "./scripts/post-plan_updated.sh"
    pre_apply  = "./scripts/pre-apply_updated.sh"
    post_apply = "./scripts/post-apply_updated.sh"
  }
}`)
}

func testAccScalrWorkspaceUpdateWorkingDir(rInt int) string {
	return fmt.Sprintf(testAccScalrWorkspaceCommonConfig, rInt, defaultAccount, `
resource "scalr_workspace" "test" {
  name                  = "workspace-updated"
  environment_id 		= scalr_environment.test.id
  auto_apply            = false
  operations            = false
  terraform_version     = "0.12.19"
  working_directory     = ""
}`)
}
