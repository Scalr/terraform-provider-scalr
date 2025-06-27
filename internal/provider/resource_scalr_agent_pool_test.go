package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrAgentPool_basic(t *testing.T) {
	pool := &scalr.AgentPool{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAgentPoolDestroy,
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

func TestAccScalrAgentPool_update(t *testing.T) {
	pool := &scalr.AgentPool{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAgentPoolDestroy,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAgentPoolDestroy,
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
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

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

func testAccCheckScalrAgentPoolDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

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
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount, rInt)
}

func testAccScalrAgentPoolUpdate() string {
	return `
resource "scalr_agent_pool" "test" {
  name           = "agent_pool-updated"
}`
}
