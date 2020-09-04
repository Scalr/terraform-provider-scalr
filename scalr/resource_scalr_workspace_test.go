package scalr

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccScalrWorkspace_basic(t *testing.T) {
	workspace := &scalr.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.foobar", workspace),
					testAccCheckScalrWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "working_directory", ""),
					resource.TestCheckResourceAttrSet("scalr_workspace.foobar", "created_by.0.full_name"),
					resource.TestCheckResourceAttrSet("scalr_workspace.foobar", "created_by.0.email"),
					resource.TestCheckResourceAttrSet("scalr_workspace.foobar", "created_by.0.username"),
				),
			},
		},
	})
}

func TestAccScalrWorkspace_monorepo(t *testing.T) {
	workspace := &scalr.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspace_monorepo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.foobar", workspace),
					testAccCheckScalrWorkspaceMonorepoAttributes(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "name", "workspace-monorepo"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "working_directory", "/db"),
				),
			},
		},
	})
}

func TestAccScalrWorkspace_renamed(t *testing.T) {
	workspace := &scalr.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.foobar", workspace),
					testAccCheckScalrWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "working_directory", ""),
				),
			},

			{
				PreConfig: testAccCheckScalrWorkspaceRename,
				Config:    testAccScalrWorkspace_renamed,
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.foobar", workspace),
					testAccCheckScalrWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "working_directory", ""),
				),
			},
		},
	})
}
func TestAccScalrWorkspace_update(t *testing.T) {
	workspace := &scalr.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.foobar", workspace),
					testAccCheckScalrWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "working_directory", ""),
				),
			},

			{
				Config: testAccScalrWorkspace_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkspaceExists(
						"scalr_workspace.foobar", workspace),
					testAccCheckScalrWorkspaceAttributesUpdated(workspace),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "operations", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "queue_all_runs", "false"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "terraform_version", "0.12.19"),
					resource.TestCheckResourceAttr(
						"scalr_workspace.foobar", "working_directory", "terraform/test"),
				),
			},
		},
	})
}

func TestAccScalrWorkspace_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspace_basic,
			},

			{
				ResourceName:      "scalr_workspace.foobar",
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

		if workspace.QueueAllRuns != true {
			return fmt.Errorf("Bad queue all runs: %t", workspace.QueueAllRuns)
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

func testAccCheckScalrWorkspaceRename() {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	w, err := scalrClient.Workspaces.Update(
		context.Background(),
		"existing-env",
		"workspace-test",
		scalr.WorkspaceUpdateOptions{Name: scalr.String("renamed-out-of-band")},
	)
	if err != nil {
		log.Fatalf("Could not rename the workspace out of band: %v", err)
	}

	if w.Name != "renamed-out-of-band" {
		log.Fatalf("Failed to rename the workspace out of band: %v", err)
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

		if workspace.QueueAllRuns != false {
			return fmt.Errorf("Bad queue all runs: %t", workspace.QueueAllRuns)
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

const testAccScalrWorkspace_basic = `
resource "scalr_workspace" "foobar" {
  name           = "workspace-test"
  environment_id = "existing-env"
  auto_apply     = true
}`

const testAccScalrWorkspace_monorepo = `
resource "scalr_workspace" "foobar" {
  name                  = "workspace-monorepo"
  environment_id        = "existing-env"
  working_directory     = "/db"
}`

const testAccScalrWorkspace_renamed = `
resource "scalr_workspace" "foobar" {
  name           = "renamed-out-of-band"
  environment_id = "existing-env"
  auto_apply     = true
}`

const testAccScalrWorkspace_update = `
resource "scalr_workspace" "foobar" {
  name                  = "workspace-updated"
  environment_id        = "existing-env"
  auto_apply            = false
  operations            = false
  queue_all_runs        = false
  terraform_version     = "0.12.19"
  working_directory     = "terraform/test"
}`
