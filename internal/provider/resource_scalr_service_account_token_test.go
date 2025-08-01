package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrServiceAccountToken_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrServiceAccountTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrServiceAccountTokenBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_service_account_token.test", "description", fmt.Sprintf("desc-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account_token.test", "name", fmt.Sprintf("name-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account_token.test", "expires_in", "42",
					),
					resource.TestCheckResourceAttrPair(
						"scalr_service_account_token.test", "service_account_id",
						"scalr_service_account.test", "id",
					),
					resource.TestCheckResourceAttrSet("scalr_service_account_token.test", "token"),
				),
			},
		},
	})
}

func TestAccScalrServiceAccountToken_update(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrServiceAccountTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrServiceAccountTokenBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_service_account_token.test", "description", fmt.Sprintf("desc-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account_token.test", "name", fmt.Sprintf("name-%d", rInt),
					),
				),
			},
			{
				Config: testAccScalrServiceAccountTokenUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_service_account_token.test", "description", fmt.Sprintf("desc-updated-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"scalr_service_account_token.test", "name", fmt.Sprintf("name-updated-%d", rInt),
					),
				),
			},
		},
	})
}

func testAccCheckScalrServiceAccountTokenDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_service_account_token" {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.AccessTokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Service account token %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrServiceAccountTokenBasicConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_service_account test {
  name = "test-sa-%d"
}
resource scalr_service_account_token test {
  service_account_id = scalr_service_account.test.id
  description        = "desc-%[1]d"
  name               = "name-%[1]d"
  expires_in         = 42
}`, rInt)
}

func testAccScalrServiceAccountTokenUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_service_account test {
  name = "test-sa-%d"
}
resource scalr_service_account_token test {
  service_account_id = scalr_service_account.test.id
  description        = "desc-updated-%[1]d"
  name               = "name-updated-%[1]d"
}`, rInt)
}
