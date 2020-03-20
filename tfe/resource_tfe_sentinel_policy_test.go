package tfe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	tfe "github.com/scalr/go-tfe"
)

func TestAccTFESentinelPolicy_basic(t *testing.T) {
	policy := &tfe.Policy{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESentinelPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelPolicy_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelPolicyExists(
						"scalr_sentinel_policy.foobar", policy),
					testAccCheckTFESentinelPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "policy", "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "enforce_mode", "hard-mandatory"),
				),
			},
		},
	})
}

func TestAccTFESentinelPolicy_update(t *testing.T) {
	policy := &tfe.Policy{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESentinelPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelPolicy_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelPolicyExists(
						"scalr_sentinel_policy.foobar", policy),
					testAccCheckTFESentinelPolicyAttributes(policy),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "description", "A test policy"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "policy", "main = rule { true }"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "enforce_mode", "hard-mandatory"),
				),
			},

			{
				Config: testAccTFESentinelPolicy_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESentinelPolicyExists(
						"scalr_sentinel_policy.foobar", policy),
					testAccCheckTFESentinelPolicyAttributesUpdated(policy),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "name", "policy-test"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "description", "An updated test policy"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "policy", "main = rule { false }"),
					resource.TestCheckResourceAttr(
						"scalr_sentinel_policy.foobar", "enforce_mode", "soft-mandatory"),
				),
			},
		},
	})
}

func TestAccTFESentinelPolicy_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESentinelPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESentinelPolicy_basic,
			},

			{
				ResourceName:        "scalr_sentinel_policy.foobar",
				ImportState:         true,
				ImportStateIdPrefix: "tst-terraform/",
				ImportStateVerify:   true,
			},
		},
	})
}

func testAccCheckTFESentinelPolicyExists(
	n string, policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		p, err := tfeClient.Policies.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if p.ID != rs.Primary.ID {
			return fmt.Errorf("SentinelPolicy not found")
		}

		*policy = *p

		return nil
	}
}

func testAccCheckTFESentinelPolicyAttributes(
	policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		if policy.Enforce[0].Mode != "hard-mandatory" {
			return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
		}

		return nil
	}
}

func testAccCheckTFESentinelPolicyAttributesUpdated(
	policy *tfe.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if policy.Name != "policy-test" {
			return fmt.Errorf("Bad name: %s", policy.Name)
		}

		if policy.Enforce[0].Mode != "soft-mandatory" {
			return fmt.Errorf("Bad enforce mode: %s", policy.Enforce[0].Mode)
		}

		return nil
	}
}

func testAccCheckTFESentinelPolicyDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_sentinel_policy" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.Policies.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Sentinel policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFESentinelPolicy_basic = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_sentinel_policy" "foobar" {
  name         = "policy-test"
  description  = "A test policy"
  organization = "${scalr_organization.foobar.id}"
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}`

const testAccTFESentinelPolicy_update = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_sentinel_policy" "foobar" {
  name         = "policy-test"
  description  = "An updated test policy"
  organization = "${scalr_organization.foobar.id}"
  policy       = "main = rule { false }"
  enforce_mode = "soft-mandatory"
}`
