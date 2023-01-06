package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccScalrServiceAccountDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrServiceAccountDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_service_account.test", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_service_account.test", "email",
						"scalr_service_account.test", "email",
					),
					resource.TestCheckResourceAttr("data.scalr_service_account.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("data.scalr_service_account.test", "created_by.#", "1"),
				),
			},
		},
	})
}

func testAccScalrServiceAccountDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_service_account test {
  name       = "test-sa-%d"
}

data scalr_service_account test {
  email      = scalr_service_account.test.email
}`, rInt)
}
