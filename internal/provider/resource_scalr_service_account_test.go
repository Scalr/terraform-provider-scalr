package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrServiceAccount_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrServiceAccountBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_service_account.test", "id"),
					resource.TestCheckResourceAttr(
						"scalr_service_account.test", "name", fmt.Sprintf("test-sa-%d", rInt),
					),
					resource.TestCheckResourceAttrSet("scalr_service_account.test", "email"),
					resource.TestCheckResourceAttr(
						"scalr_service_account.test", "description", fmt.Sprintf("desc-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account.test", "status", string(scalr.ServiceAccountStatusActive),
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account.test", "account_id", defaultAccount,
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account.test", "created_by.#", "1",
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account.test", "owners.#", "1",
					),
				),
			},
		},
	})
}

func TestAccScalrServiceAccount_import(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrServiceAccountBasic(rInt),
			},
			{
				ResourceName:      "scalr_service_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalrServiceAccount_update(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrServiceAccountBasic(rInt),
				Check: resource.TestCheckResourceAttr(
					"scalr_service_account.test", "name", fmt.Sprintf("test-sa-%d", rInt),
				),
			},
			{
				Config: testAccScalrServiceAccountUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_service_account.test",
						"description",
						fmt.Sprintf("desc-updated-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account.test", "status", string(scalr.ServiceAccountStatusInactive),
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account.test", "owners.#", "0",
					),
				),
			},
		},
	})
}

func testAccScalrServiceAccountBasic(rInt int) string {
	return fmt.Sprintf(`
resource scalr_service_account test {
  name        = "test-sa-%d"
  description = "desc-%[1]d"
  status      = "%[2]s"
  owners      = [scalr_iam_team.test.id]
}

resource "scalr_iam_team" "test" {
	name        = "test-%[1]d-owner"
	description = "Test team"
	users       = []
}`, rInt, scalr.ServiceAccountStatusActive)
}

func testAccScalrServiceAccountUpdate(rInt int) string {
	return fmt.Sprintf(`
resource scalr_service_account test {
  name        = "test-sa-%d"
  description = "desc-updated-%[1]d"
  status      = "%[2]s"
  owners      = []
}`, rInt, scalr.ServiceAccountStatusInactive)
}

func testAccCheckScalrServiceAccountDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_service_account" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.ServiceAccounts.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Service account %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
