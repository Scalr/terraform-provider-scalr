package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccScalrTagDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_tag test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_tag test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_tag test {name = ""}`,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrTagDataSourceByIDConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", fmt.Sprintf("test-tag-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
			{
				Config: testAccScalrTagDataSourceByNameConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", fmt.Sprintf("test-tag-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
			{
				Config: testAccScalrTagDataSourceByIDAndNameConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", fmt.Sprintf("test-tag-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func testAccScalrTagDataSourceByIDConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "test-tag-%[1]d"
  account_id = "%[2]s"
}

data scalr_tag test {
  id         = scalr_tag.test.id
  account_id = "%[2]s"
}`, rInt, defaultAccount)
}

func testAccScalrTagDataSourceByNameConfig(rInt int) string {
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

func testAccScalrTagDataSourceByIDAndNameConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "test-tag-%[1]d"
  account_id = "%[2]s"
}

data scalr_tag test {
  id         = scalr_tag.test.id
  name       = scalr_tag.test.name
  account_id = "%[2]s"
}`, rInt, defaultAccount)
}
