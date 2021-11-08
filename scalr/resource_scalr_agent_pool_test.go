package scalr

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccScalrAgentPool_basic(t *testing.T) {
	pool := &scalr.AgentPool{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAgentPoolExists("scalr_agent_pool.test", pool),
					resource.TestCheckResourceAttr(
						"scalr_agent_pool.test", "name", fmt.Sprintf("agent_pool-test-%d", rInt),
					),
					resource.TestCheckResourceAttr("scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func TestAccScalrAgentPool_renamed(t *testing.T) {
	pool := &scalr.AgentPool{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAgentPoolExists("scalr_agent_pool.test", pool),
					resource.TestCheckResourceAttr(
						"scalr_agent_pool.test", "name", fmt.Sprintf("agent_pool-test-%d", rInt),
					),
					resource.TestCheckResourceAttr("scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},

			{
				PreConfig: testAccCheckScalrAgentPoolRename(pool),
				Config:    testAccScalrAgentPoolRenamed(rInt),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_agent_pool.test", "name", "renamed-outside-of-terraform"),
					resource.TestCheckResourceAttr("scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},
		},
	})
}
func TestAccScalrAgentPool_update(t *testing.T) {
	pool := &scalr.AgentPool{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_agent_pool.test", "name", fmt.Sprintf("agent_pool-test-%d", rInt),
					),
					resource.TestCheckResourceAttr("scalr_agent_pool.test", "account_id", defaultAccount),
				),
			},

			{
				Config: testAccScalrAgentPoolUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAgentPoolExists("scalr_agent_pool.test", pool),
					resource.TestCheckResourceAttr("scalr_agent_pool.test", "name", "agent_pool-updated"),
				),
			},
		},
	})
}

func TestAccScalrAgentPool_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrAgentPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAgentPoolBasic(rInt),
			},

			{
				ResourceName:      "scalr_agent_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrAgentPoolExists(resId string, pool *scalr.AgentPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the agent_pool
		r, err := scalrClient.AgentPools.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*pool = *r

		return nil
	}
}

func testAccCheckScalrAgentPoolRename(pool *scalr.AgentPool) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		r, err := scalrClient.AgentPools.Read(ctx, pool.ID)

		if err != nil {
			log.Fatalf("Error retrieving agent pool: %v", err)
		}

		r, err = scalrClient.AgentPools.Update(
			context.Background(),
			r.ID,
			scalr.AgentPoolUpdateOptions{Name: scalr.String("renamed-outside-of-terraform")},
		)
		if err != nil {
			log.Fatalf("Could not rename the agent pool outside of terraform: %v", err)
		}

		if r.Name != "renamed-outside-of-terraform" {
			log.Fatalf("Failed to rename the agent pool outside of terraform: %v", err)
		}
	}
}

func testAccCheckScalrAgentPoolDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_agent_pool" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.AgentPools.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("AgentPool %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrAgentPoolBasic(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name           = "agent_pool-test-%d"
  account_id     = "%s"

}

resource "scalr_agent_pool" "test" {
  name           = "agent_pool-test-%d"
  account_id     = "%s"
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount, rInt, defaultAccount)
}

func testAccScalrAgentPoolRenamed(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name           = "agent_pool-test-%d"
  account_id     = "%s"

}
resource "scalr_agent_pool" "test" {
  name           = "renamed-outside-of-terraform"
  account_id     = "%s"
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount, defaultAccount)
}

func testAccScalrAgentPoolUpdate() string {
	return fmt.Sprintf(`
resource "scalr_agent_pool" "test" {
  name           = "agent_pool-updated"
  account_id     = "%s"
}`, defaultAccount)
}
