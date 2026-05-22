package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccScalrWorkspaceVarSet_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWorkspaceVarSetDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWorkspaceVarSetBasic(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrWorkspaceVarSetExists("scalr_workspace_var_set.test"),
						resource.TestCheckResourceAttrSet("scalr_workspace_var_set.test", "workspace_id"),
						resource.TestCheckResourceAttrSet("scalr_workspace_var_set.test", "var_set_id"),
						resource.TestCheckResourceAttrPair(
							"scalr_workspace_var_set.test", "workspace_id",
							"scalr_workspace.test", "id",
						),
						resource.TestCheckResourceAttrPair(
							"scalr_workspace_var_set.test", "var_set_id",
							"scalr_var_set.test", "id",
						),
					),
				},
			},
		},
	)
}

func TestAccScalrWorkspaceVarSet_multiple(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWorkspaceVarSetDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWorkspaceVarSetMultiple(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrWorkspaceVarSetExists("scalr_workspace_var_set.link1"),
						testAccCheckScalrWorkspaceVarSetExists("scalr_workspace_var_set.link2"),
					),
				},
			},
		},
	)
}

func TestAccScalrWorkspaceVarSet_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWorkspaceVarSetDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWorkspaceVarSetBasic(rInt),
				},
				{
					ResourceName:      "scalr_workspace_var_set.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		},
	)
}

func testAccCheckScalrWorkspaceVarSetExists(resID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resID]
		if !ok {
			return fmt.Errorf("not found: %s", resID)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		parts := strings.SplitN(rs.Primary.ID, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid ID format: %s", rs.Primary.ID)
		}
		workspaceID, varSetID := parts[0], parts[1]

		varSets, err := scalrClient.Workspace.ListWorkspaceVariableSets(ctx, workspaceID, nil)
		if err != nil {
			return fmt.Errorf("error listing variable sets for workspace %s: %w", workspaceID, err)
		}

		for _, vs := range varSets {
			if vs.ID == varSetID {
				return nil
			}
		}

		return fmt.Errorf("variable set %s not linked to workspace %s", varSetID, workspaceID)
	}
}

func testAccCheckScalrWorkspaceVarSetDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_workspace_var_set" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		parts := strings.SplitN(rs.Primary.ID, "/", 2)
		if len(parts) != 2 {
			continue
		}
		workspaceID, varSetID := parts[0], parts[1]

		varSets, err := scalrClient.Workspace.ListWorkspaceVariableSets(ctx, workspaceID, nil)
		if err != nil {
			// Workspace may already be deleted — treat as success.
			continue
		}

		for _, vs := range varSets {
			if vs.ID == varSetID {
				return fmt.Errorf("variable set %s still linked to workspace %s", varSetID, workspaceID)
			}
		}
	}

	return nil
}

func testAccScalrWorkspaceVarSetBasic(rInt int) string {
	return fmt.Sprintf(
		`
resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "%s"
}

resource "scalr_workspace" "test" {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource "scalr_var_set" "test" {
  name         = "test-var-set-%[1]d"
  environments = ["*"]
}

resource "scalr_workspace_var_set" "test" {
  workspace_id = scalr_workspace.test.id
  var_set_id   = scalr_var_set.test.id
}
`, rInt, defaultAccount,
	)
}

func testAccScalrWorkspaceVarSetMultiple(rInt int) string {
	return fmt.Sprintf(
		`
resource "scalr_environment" "test" {
  name       = "test-env-%d"
  account_id = "%s"
}

resource "scalr_workspace" "test" {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource "scalr_var_set" "vs1" {
  name = "test-var-set-%[1]d-1"
  environments = ["*"]
}

resource "scalr_var_set" "vs2" {
  name = "test-var-set-%[1]d-2"
  environments = [scalr_environment.test.id]
}

resource "scalr_workspace_var_set" "link1" {
  workspace_id = scalr_workspace.test.id
  var_set_id   = scalr_var_set.vs1.id
}

resource "scalr_workspace_var_set" "link2" {
  workspace_id = scalr_workspace.test.id
  var_set_id   = scalr_var_set.vs2.id
}
`, rInt, defaultAccount,
	)
}
