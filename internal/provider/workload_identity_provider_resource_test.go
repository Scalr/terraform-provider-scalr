package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrWorkloadIdentityProvider_basic(t *testing.T) {
	providerName := acctest.RandomWithPrefix("test-wip")
	providerURL := "https://example.com"
	allowedAudience := "myAud"
	provider := &scalr.WorkloadIdentityProvider{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkloadIdentityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkloadIdentityProviderBasic(providerName, providerURL, allowedAudience),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkloadIdentityProviderExists("scalr_workload_identity_provider.test", provider),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "name", providerName),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "url", providerURL),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "allowed_audiences.#", "1"),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "allowed_audiences.0", allowedAudience),
				),
			},
		},
	})
}

func TestAccScalrWorkloadIdentityProvider_import(t *testing.T) {
	providerName := acctest.RandomWithPrefix("test-wip")
	providerURL := "https://example.com"
	allowedAudience := "myAud"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkloadIdentityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkloadIdentityProviderBasic(providerName, providerURL, allowedAudience),
			},
			{
				ResourceName:      "scalr_workload_identity_provider.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalrWorkloadIdentityProvider_update(t *testing.T) {
	providerName := acctest.RandomWithPrefix("test-wip")
	providerNameUpdated := acctest.RandomWithPrefix("test-wip-updated")
	providerURL := "https://example.com"
	allowedAudience := "myAud"
	allowedAudienceUpdated := "myAudNew"
	provider := &scalr.WorkloadIdentityProvider{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrWorkloadIdentityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkloadIdentityProviderBasic(providerName, providerURL, allowedAudience),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkloadIdentityProviderExists("scalr_workload_identity_provider.test", provider),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "name", providerName),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "url", providerURL),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "allowed_audiences.#", "1"),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "allowed_audiences.0", allowedAudience),
				),
			},
			{
				Config: testAccScalrWorkloadIdentityProviderBasic(providerNameUpdated, providerURL, allowedAudienceUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrWorkloadIdentityProviderExists("scalr_workload_identity_provider.test", provider),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "name", providerNameUpdated),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "url", providerURL),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "allowed_audiences.#", "1"),
					resource.TestCheckResourceAttr("scalr_workload_identity_provider.test", "allowed_audiences.0", allowedAudienceUpdated),
				),
			},
		},
	})
}

func testAccScalrWorkloadIdentityProviderBasic(name, url, allowedAudience string) string {
	return fmt.Sprintf(`
resource "scalr_workload_identity_provider" "test" {
  name              = "%s"
  url               = "%s"
  allowed_audiences = ["%s"]
}`, name, url, allowedAudience)
}

func testAccCheckScalrWorkloadIdentityProviderExists(resId string, provider *scalr.WorkloadIdentityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		p, err := scalrClient.WorkloadIdentityProviders.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*provider = *p

		return nil
	}
}

func testAccCheckScalrWorkloadIdentityProviderDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_workload_identity_provider" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.WorkloadIdentityProviders.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Workload Identity Provider %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
