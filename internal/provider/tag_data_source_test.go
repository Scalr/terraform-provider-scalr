package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrTagDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_tag test {}`,
				ExpectError: regexp.MustCompile("At least one of these attributes must be configured: \\[id,name]"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_tag test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_tag test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
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
