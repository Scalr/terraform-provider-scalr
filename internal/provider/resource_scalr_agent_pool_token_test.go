package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrAgentPoolToken_basic(t *testing.T) {
	token := &scalr.AccessToken{}

	var pool scalr.AgentPool
	if isAccTest() {
		pool = createPool(t)
		defer deletePool(t, pool)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAgentPoolTokenDestroy,
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

func TestAccScalrAgentPoolToken_update(t *testing.T) {
	var pool scalr.AgentPool
	if isAccTest() {
		pool = createPool(t)
		defer deletePool(t, pool)
	}
	token := &scalr.AccessToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAgentPoolTokenDestroy,
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

func testAccCheckScalrAgentPoolTokenExists(resId string, pool scalr.AgentPool, token *scalr.AccessToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the token
		l, err := scalrClient.AgentPoolTokens.List(ctx, pool.ID, scalr.AccessTokenListOptions{})
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
		Name:    ptr(name),
		Account: &scalr.Account{ID: defaultAccount},
	})

	if err != nil {
		t.Fatalf("Unable to create an agent pool: %s", err)
	}
	return *r
}

func testAccCheckScalrAgentPoolTokenDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_agent_pool_token" {
			continue
		}

		poolID := rs.Primary.Attributes["agent_pool_id"]

		// the agent pool must be deleted along with token
		l, _ := scalrClient.AgentPoolTokens.List(ctx, poolID, scalr.AccessTokenListOptions{})
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

func testAccScalrAgentPoolTokenUpdate(pool scalr.AgentPool) string {
	return fmt.Sprintf(`
resource "scalr_agent_pool_token" "test" {
  description           = "agent_pool_token-updated"
  agent_pool_id     = "%s"
}`, pool.ID)
}
