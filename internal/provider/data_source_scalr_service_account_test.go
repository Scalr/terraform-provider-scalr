package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scalr/go-scalr"
)

func TestAccScalrServiceAccountDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_service_account test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `email,id` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_service_account test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_service_account test {email = ""}`,
				ExpectError: regexp.MustCompile("expected \"email\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrServiceAccountDataSourceByIDConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_service_account.test", "id"),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "name", fmt.Sprintf("test-sa-%d", rInt),
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_service_account.test", "email",
						"scalr_service_account.test", "email",
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "description", fmt.Sprintf("desc-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "account_id", defaultAccount,
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "created_by.#", "1",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_service_account.test", "owners",
						"scalr_service_account.test", "owners",
					),
				),
			},
			{
				Config: testAccScalrServiceAccountDataSourceByEmailConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_service_account.test", "id",
						"scalr_service_account.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "name", fmt.Sprintf("test-sa-%d", rInt),
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_service_account.test", "email",
						"scalr_service_account.test", "email",
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "description", fmt.Sprintf("desc-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "account_id", defaultAccount,
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "created_by.#", "1",
					),
				),
			},
			{
				Config: testAccScalrServiceAccountDataSourceByIDAndEmailConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_service_account.test", "id",
						"scalr_service_account.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "name", fmt.Sprintf("test-sa-%d", rInt),
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_service_account.test", "email",
						"scalr_service_account.test", "email",
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "description", fmt.Sprintf("desc-%d", rInt),
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "account_id", defaultAccount,
					),
					resource.TestCheckResourceAttr(
						"data.scalr_service_account.test", "created_by.#", "1",
					),
				),
			},
		},
	})
}

func testAccScalrServiceAccountDataSourceByIDConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_service_account test {
  name        = "test-sa-%d"
  description = "desc-%[1]d"
  status      = "%[2]s"
}

data scalr_service_account test {
  id = scalr_service_account.test.id
}`, rInt, scalr.ServiceAccountStatusActive)
}

func testAccScalrServiceAccountDataSourceByEmailConfig(rInt int) string {
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
}

data scalr_service_account test {
  email = scalr_service_account.test.email
}`, rInt, scalr.ServiceAccountStatusInactive)
}

func testAccScalrServiceAccountDataSourceByIDAndEmailConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_service_account test {
  name        = "test-sa-%d"
  description = "desc-%[1]d"
  status      = "%[2]s"
}

data scalr_service_account test {
  id    = scalr_service_account.test.id
  email = scalr_service_account.test.email
}`, rInt, scalr.ServiceAccountStatusInactive)
}
