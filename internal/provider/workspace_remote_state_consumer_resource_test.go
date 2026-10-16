package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccScalrWorkspaceRemoteStateConsumer_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWorkspaceRemoteStateConsumerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWorkspaceRemoteStateConsumerConfig(rInt, true),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrWorkspaceRemoteStateConsumerExists("scalr_workspace_remote_state_consumer.child1"),
						testAccCheckScalrWorkspaceRemoteStateConsumerExists("scalr_workspace_remote_state_consumer.child2"),
						resource.TestCheckResourceAttrPair(
							"scalr_workspace_remote_state_consumer.child1", "workspace_id",
							"scalr_workspace.parent", "id",
						),
						resource.TestCheckResourceAttrPair(
							"scalr_workspace_remote_state_consumer.child1", "consumer_id",
							"scalr_workspace.child1", "id",
						),
						resource.TestCheckResourceAttr("scalr_workspace.parent", "remote_state_sharing", "false"),
						testAccCheckScalrWorkspaceStateSharing("scalr_workspace.parent", false),
					),
				},
				{
					// Refresh picks up the consumers added by the standalone resources
					// without producing a diff on the parent workspace.
					Config: testAccScalrWorkspaceRemoteStateConsumerConfig(rInt, true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("scalr_workspace.parent", "remote_state_consumers.#", "2"),
					),
				},
				{
					Config: testAccScalrWorkspaceRemoteStateConsumerConfig(rInt, false),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrWorkspaceRemoteStateConsumerExists("scalr_workspace_remote_state_consumer.child1"),
						testAccCheckScalrWorkspaceRemoteStateConsumerDestroy,
					),
				},
				{
					ResourceName:      "scalr_workspace_remote_state_consumer.child1",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		},
	)
}

func TestAccScalrWorkspaceRemoteStateConsumer_sharedWorkspace(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrWorkspaceRemoteStateConsumerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrWorkspaceRemoteStateConsumerSharedConfig(rInt),
					ExpectError: regexp.MustCompile(
						"not allowed when the state is shared with environment",
					),
				},
			},
		},
	)
}

func testAccCheckScalrWorkspaceRemoteStateConsumerExists(resID string) resource.TestCheckFunc {
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
		workspaceID, consumerID := parts[0], parts[1]

		consumers, err := scalrClient.Workspace.ListRemoteStateConsumers(ctx, workspaceID, nil)
		if err != nil {
			return fmt.Errorf("error listing remote state consumers for workspace %s: %w", workspaceID, err)
		}

		for _, c := range consumers {
			if c.ID == consumerID {
				return nil
			}
		}

		return fmt.Errorf("workspace %s is not a remote state consumer of workspace %s", consumerID, workspaceID)
	}
}

func testAccCheckScalrWorkspaceRemoteStateConsumerDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	// Collect the consumers still managed in the state, to support partial destroy checks.
	managed := make(map[string]bool)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "scalr_workspace_remote_state_consumer" {
			managed[rs.Primary.ID] = true
		}
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_workspace" {
			continue
		}

		consumers, err := scalrClient.Workspace.ListRemoteStateConsumers(ctx, rs.Primary.ID, nil)
		if err != nil {
			// Workspace may already be deleted — treat as success.
			continue
		}

		for _, c := range consumers {
			if !managed[rs.Primary.ID+"/"+c.ID] {
				return fmt.Errorf("workspace %s is still a remote state consumer of workspace %s", c.ID, rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccScalrWorkspaceRemoteStateConsumerConfig(rInt int, withChild2 bool) string {
	child2 := ""
	if withChild2 {
		child2 = `
resource "scalr_workspace_remote_state_consumer" "child2" {
  workspace_id = scalr_workspace.parent.id
  consumer_id  = scalr_workspace.child2.id
}
`
	}
	return fmt.Sprintf(
		`
resource "scalr_environment" "test" {
  name       = "test-env-%[1]d"
  account_id = "%[2]s"
}

resource "scalr_workspace" "parent" {
  name                 = "parent-%[1]d"
  environment_id       = scalr_environment.test.id
  remote_state_sharing = false
}

resource "scalr_workspace" "child1" {
  name           = "child1-%[1]d"
  environment_id = scalr_environment.test.id
}

resource "scalr_workspace" "child2" {
  name           = "child2-%[1]d"
  environment_id = scalr_environment.test.id
}

resource "scalr_workspace_remote_state_consumer" "child1" {
  workspace_id = scalr_workspace.parent.id
  consumer_id  = scalr_workspace.child1.id
}
%[3]s`, rInt, defaultAccount, child2,
	)
}

func testAccScalrWorkspaceRemoteStateConsumerSharedConfig(rInt int) string {
	return fmt.Sprintf(
		`
resource "scalr_environment" "test" {
  name       = "test-env-%[1]d"
  account_id = "%[2]s"
}

resource "scalr_workspace" "parent" {
  name                 = "parent-%[1]d"
  environment_id       = scalr_environment.test.id
  remote_state_sharing = true
}

resource "scalr_workspace" "child" {
  name           = "child-%[1]d"
  environment_id = scalr_environment.test.id
}

resource "scalr_workspace_remote_state_consumer" "child" {
  workspace_id = scalr_workspace.parent.id
  consumer_id  = scalr_workspace.child.id
}
`, rInt, defaultAccount,
	)
}
