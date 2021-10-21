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

func TestAccScalrAgentPoolToken_basic(t *testing.T) {
	token := &scalr.AgentPoolToken{}

	var pool scalr.AgentPool
	if isAccTest() {
		pool = createPool(t)
		defer deletePool(t, pool)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrAgentPoolTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolTokenBasic(pool),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAgentPoolTokenExists("scalr_agent_pool_token.test", pool, token),
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "description", "agent_pool_token-test"),
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "agent_pool_id", pool.ID),
				),
			},
		},
	})
}

func TestAccScalrAgentPoolToken_changed_outside(t *testing.T) {

	var pool scalr.AgentPool
	if isAccTest() {
		pool = createPool(t)
		defer deletePool(t, pool)
	}
	token := &scalr.AgentPoolToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrAgentPoolTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolTokenBasic(pool),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAgentPoolTokenExists("scalr_agent_pool_token.test", pool, token),
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "description", "agent_pool_token-test"),
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "agent_pool_id", pool.ID),
				),
			},

			{
				PreConfig: testAccCheckScalrAgentPoolTokenChangedOutside(pool, token),
				Config:    testAccScalrAgentPoolTokenChangedOutside(pool),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "description", "changed-outside-of-terraform"),
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "agent_pool_id", pool.ID),
				),
			},
		},
	})
}
func TestAccScalrAgentPoolToken_update(t *testing.T) {
	var pool scalr.AgentPool
	if isAccTest() {
		pool = createPool(t)
		defer deletePool(t, pool)
	}
	token := &scalr.AgentPoolToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrAgentPoolTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolTokenBasic(pool),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "description", "agent_pool_token-test"),
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "agent_pool_id", pool.ID),
				),
			},

			{
				Config: testAccScalrAgentPoolTokenUpdate(pool),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAgentPoolTokenExists("scalr_agent_pool_token.test", pool, token),
					resource.TestCheckResourceAttr("scalr_agent_pool_token.test", "description", "agent_pool_token-updated"),
				),
			},
		},
	})
}

func testAccCheckScalrAgentPoolTokenExists(resId string, pool scalr.AgentPool, token *scalr.AgentPoolToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the token
		l, err := scalrClient.AgentPoolTokens.List(ctx, pool.ID, scalr.AgentPoolTokenListOptions{})
		if err != nil {
			return err
		}

		if len(l.Items) != 1 {
			return fmt.Errorf("There are more than one token for pool: %d", len(l.Items))
		}
		if l.Items[0].ID != rs.Primary.ID {
			return fmt.Errorf("Expected to get: %s, got: %s.", rs.Primary.ID, l.Items[0].ID)
		}

		*token = *l.Items[0]

		return nil
	}
}

func testAccCheckScalrAgentPoolTokenChangedOutside(pool scalr.AgentPool, token *scalr.AgentPoolToken) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		r, err := scalrClient.AccessTokens.Update(
			context.Background(),
			token.ID,
			scalr.AccessTokenUpdateOptions{Description: scalr.String("changed-outside-of-terraform")},
		)
		if err != nil {
			log.Fatalf("Could not update the agent pool outside of terraform: %v", err)
		}

		if r.Description != "changed-outside-of-terraform" {
			log.Fatalf("Failed to update the agent pool outside of terraform: %v", err)
		}
	}
}

func deletePool(t *testing.T, pool scalr.AgentPool) {
	scalrClient, err := createScalrClient()
	if err != nil {
		t.Fatalf("Unable to create a Scalr client: %s", err)
	}

	err = scalrClient.AgentPools.Delete(ctx, pool.ID)
	if err != nil {
		t.Fatalf("Unable to delete an agent pool: %s", err)
	}
}

func createPool(t *testing.T) scalr.AgentPool {
	name := fmt.Sprintf("provider-test-pool-%d", GetRandomInteger())

	scalrClient, err := createScalrClient()
	if err != nil {
		t.Fatalf("Unable to create a Scalr client: %s", err)
	}

	r, err := scalrClient.AgentPools.Create(ctx, scalr.AgentPoolCreateOptions{
		Name:    scalr.String(name),
		Account: &scalr.Account{ID: defaultAccount},
	})

	if err != nil {
		t.Fatalf("Unable to create an agent pool: %s", err)
	}
	return *r
}

func testAccCheckScalrAgentPoolTokenDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_agent_pool_token" {
			continue
		}

		poolID := rs.Primary.Attributes["agent_pool_id"]

		// the agent pool must be deleted along with token
		l, _ := scalrClient.AgentPoolTokens.List(ctx, poolID, scalr.AgentPoolTokenListOptions{})
		if len(l.Items) > 0 {
			return fmt.Errorf("AgentPoolToken %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrAgentPoolTokenBasic(pool scalr.AgentPool) string {
	return fmt.Sprintf(`

resource "scalr_agent_pool_token" "test" {
  description           = "agent_pool_token-test"
  agent_pool_id     = "%s"
}`, pool.ID)
}

func testAccScalrAgentPoolTokenChangedOutside(pool scalr.AgentPool) string {
	return fmt.Sprintf(`
resource "scalr_agent_pool_token" "test" {
  description           = "changed-outside-of-terraform"
  agent_pool_id     = "%s"
}`, pool.ID)
}

func testAccScalrAgentPoolTokenUpdate(pool scalr.AgentPool) string {
	return fmt.Sprintf(`
resource "scalr_agent_pool_token" "test" {
  description           = "agent_pool_token-updated"
  agent_pool_id     = "%s"
}`, pool.ID)
}
