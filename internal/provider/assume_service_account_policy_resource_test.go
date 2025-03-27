package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrAssumeServiceAccountPolicy_basic(t *testing.T) {
	policyName := acctest.RandomWithPrefix("test-policy")
	policy := &scalr.AssumeServiceAccountPolicy{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAssumeServiceAccountPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAssumeServiceAccountPolicyBasic(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAssumeServiceAccountPolicyExists("scalr_assume_service_account_policy.test", policy),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "name", policyName),
					resource.TestCheckResourceAttrSet("scalr_assume_service_account_policy.test", "service_account_id"),
					resource.TestCheckResourceAttrSet("scalr_assume_service_account_policy.test", "provider_id"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "maximum_session_duration", "3600"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.#", "1"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.claim", "sub"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.value", "12345"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.operator", "eq"),
				),
			},
		},
	})
}

func TestAccScalrAssumeServiceAccountPolicy_import(t *testing.T) {
	policyName := acctest.RandomWithPrefix("test-policy")
	policy := &scalr.AssumeServiceAccountPolicy{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAssumeServiceAccountPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAssumeServiceAccountPolicyBasic(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAssumeServiceAccountPolicyExists("scalr_assume_service_account_policy.test", policy),
				),
			},
			{
				ResourceName:      "scalr_assume_service_account_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					if policy == nil || policy.ServiceAccount == nil || policy.ID == "" {
						return "", fmt.Errorf("policy or its fields are nil")
					}
					return fmt.Sprintf("%s:%s", policy.ServiceAccount.ID, policy.ID), nil
				},
			},
		},
	})
}

func TestAccScalrAssumeServiceAccountPolicy_update(t *testing.T) {
	policyName := acctest.RandomWithPrefix("test-policy")
	policyNameUpdated := acctest.RandomWithPrefix("test-policy-updated")
	policy := &scalr.AssumeServiceAccountPolicy{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrAssumeServiceAccountPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAssumeServiceAccountPolicyBasic(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAssumeServiceAccountPolicyExists("scalr_assume_service_account_policy.test", policy),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "name", policyName),
					resource.TestCheckResourceAttrSet("scalr_assume_service_account_policy.test", "service_account_id"),
					resource.TestCheckResourceAttrSet("scalr_assume_service_account_policy.test", "provider_id"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "maximum_session_duration", "3600"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.#", "1"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.claim", "sub"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.value", "12345"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.operator", "eq"),
				),
			},
			{
				Config: testAccScalrAssumeServiceAccountPolicyUpdated(policyNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrAssumeServiceAccountPolicyExists("scalr_assume_service_account_policy.test", policy),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "name", policyNameUpdated),
					resource.TestCheckResourceAttrSet("scalr_assume_service_account_policy.test", "service_account_id"),
					resource.TestCheckResourceAttrSet("scalr_assume_service_account_policy.test", "provider_id"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "maximum_session_duration", "7200"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.#", "2"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.claim", "aud"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.value", "67890"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.0.operator", "eq"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.1.claim", "sub"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.1.value", "12345"),
					resource.TestCheckResourceAttr("scalr_assume_service_account_policy.test", "claim_condition.1.operator", "like"),
				),
			},
		},
	})
}

func testAccScalrAssumeServiceAccountPolicyBasic(name string) string {
	return fmt.Sprintf(`
resource "scalr_service_account" "test" {
  name = "%[1]s"
}

resource "scalr_workload_identity_provider" "test" {
  name              = "%[1]s"
  url               = "https://test.test/%[1]s"
  allowed_audiences = ["test"]
}

resource "scalr_assume_service_account_policy" "test" {
  name                     = "%[1]s"
  service_account_id       = scalr_service_account.test.id
  provider_id              = scalr_workload_identity_provider.test.id
  maximum_session_duration = 3600
  claim_condition {
    claim    = "sub"
    value    = "12345"
    operator = "eq"
  }
}`, name)
}

func testAccScalrAssumeServiceAccountPolicyUpdated(name string) string {
	return fmt.Sprintf(`
resource "scalr_service_account" "test" {
  name = "%[1]s"
}

resource "scalr_workload_identity_provider" "test" {
  name              = "%[1]s"
  url               = "https://test.test/%[1]s"
  allowed_audiences = ["test"]
}

resource "scalr_assume_service_account_policy" "test" {
  name                     = "%[1]s"
  service_account_id       = scalr_service_account.test.id
  provider_id              = scalr_workload_identity_provider.test.id
  maximum_session_duration = 7200
  claim_condition {
    claim    = "sub"
    value    = "12345"
    operator = "like"
  }
  claim_condition {
    claim    = "aud"
    value    = "67890"
  }
}`, name)
}

func testAccCheckScalrAssumeServiceAccountPolicyExists(resId string, policy *scalr.AssumeServiceAccountPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		p, err := scalrClient.AssumeServiceAccountPolicies.Read(ctx, rs.Primary.Attributes["service_account_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		*policy = *p

		return nil
	}
}

func testAccCheckScalrAssumeServiceAccountPolicyDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_assume_service_account_policy" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.AssumeServiceAccountPolicies.Read(ctx, rs.Primary.Attributes["service_account_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Assume Service Account Policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
