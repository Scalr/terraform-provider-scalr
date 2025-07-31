package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var namespaceName = acctest.RandomWithPrefix("test-namespace")

func TestAccScalrModuleNamespaceDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_module_namespace test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrModuleNamespaceDataSourceInitConfig,
			},
			{
				Config: testAccScalrModuleNamespaceDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_module_namespace.test", "scalr_module_namespace.test"),
					resource.TestCheckResourceAttrPair(
						"data.scalr_module_namespace.test", "name",
						"scalr_module_namespace.test", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_module_namespace.test", "is_shared",
						"scalr_module_namespace.test", "is_shared",
					),
					resource.TestCheckResourceAttr("data.scalr_module_namespace.test", "is_shared", "false"),
				),
			},
			{
				Config: testAccScalrModuleNamespaceDataSourceInitConfig,
			},
		},
	})
}

var testAccScalrModuleNamespaceDataSourceInitConfig = fmt.Sprintf(`
resource "scalr_module_namespace" "test" {
  name = "%s"
}
`, namespaceName)

var testAccScalrModuleNamespaceDataSourceConfig = testAccScalrModuleNamespaceDataSourceInitConfig + `
data "scalr_module_namespace" "test" {
  name = scalr_module_namespace.test.name
}
`
