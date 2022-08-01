package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccScalrTagDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrTagDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", fmt.Sprintf("test-tag-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func testAccScalrTagDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "test-tag-%[1]d"
  account_id = "%[2]s"
}

data scalr_tag test {
  name       = scalr_tag.test.name
  account_id = "%[2]s"
}`, rInt, defaultAccount)
}
