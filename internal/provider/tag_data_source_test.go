package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrTagDataSource_basic(t *testing.T) {
	tagName := acctest.RandomWithPrefix("test-tag")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_tag test {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
			},
			{
				Config:      `data scalr_tag test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data scalr_tag test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config: testAccScalrTagDataSourceByIDConfig(tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
			{
				Config: testAccScalrTagDataSourceByNameConfig(tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
			{
				Config: testAccScalrTagDataSourceByIDAndNameConfig(tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func testAccScalrTagDataSourceByIDConfig(name string) string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "%[1]s"
  account_id = "%[2]s"
}

data scalr_tag test {
  id         = scalr_tag.test.id
  account_id = "%[2]s"
}`, name, defaultAccount)
}

func testAccScalrTagDataSourceByNameConfig(name string) string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "%[1]s"
  account_id = "%[2]s"
}

data scalr_tag test {
  name       = scalr_tag.test.name
  account_id = "%[2]s"
}`, name, defaultAccount)
}

func testAccScalrTagDataSourceByIDAndNameConfig(name string) string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "%[1]s"
  account_id = "%[2]s"
}

data scalr_tag test {
  id         = scalr_tag.test.id
  name       = scalr_tag.test.name
  account_id = "%[2]s"
}`, name, defaultAccount)
}

func TestAccScalrTagDataSource_UpgradeFromSDK(t *testing.T) {
	tagName := acctest.RandomWithPrefix("test-tag")

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"scalr": {
						Source:            "registry.scalr.io/scalr/scalr",
						VersionConstraint: "<=2.2.0",
					},
				},
				Config: testAccScalrTagDataSourceByIDConfig(tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(t),
				Config:                   testAccScalrTagDataSourceByIDConfig(tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("data.scalr_tag.test", "account_id", defaultAccount),
				),
			},
		},
	})
}
