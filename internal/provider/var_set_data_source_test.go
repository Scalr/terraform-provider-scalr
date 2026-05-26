package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrVarSetDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-var-set")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			Steps: []resource.TestStep{
				{
					Config:      `data "scalr_var_set" "test" {}`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
				},
				{
					Config:      `data "scalr_var_set" "test" { id = "" }`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("Attribute id must not be empty"),
				},
				{
					Config:      `data "scalr_var_set" "test" { name = "" }`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("Attribute name must not be empty"),
				},
				{
					Config: testAccScalrVarSetDataSourceByID(name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_var_set.test", "id"),
						resource.TestCheckResourceAttr("data.scalr_var_set.test", "name", name),
						resource.TestCheckResourceAttr("data.scalr_var_set.test", "description", "test description"),
						resource.TestCheckResourceAttr("data.scalr_var_set.test", "account_id", defaultAccount),
						resource.TestCheckResourceAttrSet("data.scalr_var_set.test", "updated_at"),
						resource.TestCheckResourceAttrSet("data.scalr_var_set.test", "updated_by_email"),
						resource.TestCheckResourceAttr("data.scalr_var_set.test", "environments.#", "0"),
						resource.TestCheckResourceAttr("data.scalr_var_set.test", "owners.#", "0"),
					),
				},
				{
					Config: testAccScalrVarSetDataSourceByName(name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.scalr_var_set.test", "id"),
						resource.TestCheckResourceAttr("data.scalr_var_set.test", "name", name),
					),
				},
			},
		},
	)
}

func TestAccScalrVarSetDataSource_environments(t *testing.T) {
	name := acctest.RandomWithPrefix("test-var-set")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			Steps: []resource.TestStep{
				{
					Config: testAccScalrVarSetDataSourceSharedWithAll(name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scalr_var_set.test", "environments.#", "1"),
						resource.TestCheckTypeSetElemAttr("data.scalr_var_set.test", "environments.*", "*"),
					),
				},
			},
		},
	)
}

func testAccScalrVarSetDataSourceByID(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_var_set" "test" {
  name        = "%s"
  description = "test description"
}

data "scalr_var_set" "test" {
  id = scalr_var_set.test.id
}`, name,
	)
}

func testAccScalrVarSetDataSourceByName(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_var_set" "test" {
  name        = "%s"
  description = "test description"
}

data "scalr_var_set" "test" {
  name = scalr_var_set.test.name
}`, name,
	)
}

func testAccScalrVarSetDataSourceSharedWithAll(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_var_set" "test" {
  name         = "%s"
  environments = ["*"]
}

data "scalr_var_set" "test" {
  id = scalr_var_set.test.id
}`, name,
	)
}
