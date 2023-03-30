package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scalr/go-scalr"
	"regexp"
	"testing"
)

func TestAccScalrServiceAccountDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_service_account test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `email,id` must be specified"),
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
