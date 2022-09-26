package scalr

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccScalrVariablesDataSource(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariablesDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config: testAccScalrVariablesDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceVarsInDatasource(
						"data.scalr_variables.workspace_and_null",
						[]string{"scalr_variable.workspace_host", "scalr_variable.address", "scalr_variable.secret"},
					),
					testCheckResourceVarsInDatasource(
						"data.scalr_variables.account",
						[]string{"scalr_variable.workspace2_host", "scalr_variable.workspace_host", "scalr_variable.secret"},
					),
					testCheckResourceVarsInDatasource(
						"data.scalr_variables.workspace",
						[]string{"scalr_variable.workspace_host"},
					),
					testCheckResourceVarsInDatasource(
						"data.scalr_variables.host",
						[]string{"scalr_variable.workspace_host", "scalr_variable.workspace2_host"},
					),
					testCheckResourceVarsInDatasource(
						"data.scalr_variables.shell",
						[]string{"scalr_variable.workspace2_host", "scalr_variable.address", "scalr_variable.secret"},
					),
				),
			},
			{
				Config: testAccScalrVariablesDataSourceInitConfig,
			},
		},
	})
}

func testCheckResourceVarsInDatasource(dsName string, origNames []string) resource.TestCheckFunc {
	// check that all variable attributes in resource is equal to variables attributes in data source
	return func(s *terraform.State) error {
		ms := s.RootModule()
		if err := resource.TestCheckResourceAttr(dsName, "variables.#", strconv.Itoa(len(origNames)))(s); err != nil {
			return err
		}
		for _, variableResourceName := range origNames {
			varrs, ok := ms.Resources[variableResourceName]
			if !ok {
				return fmt.Errorf("Not found: %s in %s", variableResourceName, ms.Path)
			}
			varis := varrs.Primary
			if varis == nil {
				return fmt.Errorf("No primary instance: %s in %s", variableResourceName, ms.Path)
			}
			attr2dsKey := func(attr string) string {
				return "variables." + strconv.Itoa(schema.HashString(varis.ID)) + "." + attr
			}
			varAttrs := []string{
				"category", "hcl", "key", "sensitive", "final", "description", "workspace_id", "environment_id", "account_id",
			}
			if varis.Attributes["sensitive"] == "false" {
				varAttrs = append(varAttrs, "value")
			}

			for _, attr := range varAttrs {
				if err := resource.TestCheckResourceAttr(dsName, attr2dsKey(attr), varis.Attributes[attr])(s); err != nil {
					return fmt.Errorf("Error checking %s in data source: %v", variableResourceName, err)
				}
			}
		}
		return nil
	}
}

var testAccScalrVariablesDataSourceInitConfig = fmt.Sprintf(`
resource scalr_environment test {
	name       = "test-env-variable-data"
	account_id = "%[1]s"
  }

resource scalr_workspace test {
	name           = "test-ws-variable-data"
	environment_id = scalr_environment.test.id
}

resource scalr_workspace test2 {
	name           = "test-ws-variable-data2"
	environment_id = scalr_environment.test.id
}

resource "scalr_variable" "workspace2_host" {
	key = "host"
	value = "workspace2.scalr.com"
	category = "shell"
	hcl = false
	sensitive = false
	description = "The host of scalr workspace2."
	final = false
	workspace_id=scalr_workspace.test2.id
}

resource "scalr_variable" "workspace_host" {
	key = "host"
	value = "workspace.scalr.com"
	category = "terraform"
	hcl = false
	sensitive = false
	description = "The host of scalr workspace."
	final = true
	workspace_id=scalr_workspace.test.id
}

resource "scalr_variable" "address" {
	key = "address"
	value = "scalr.com"
	category = "shell"
	hcl = true
	sensitive = false
	description = "The address of scalr."
	final = false
}

resource "scalr_variable" "secret" {
	key = "secret"
	value = "secret-key"
	category = "shell"
	hcl = false
	sensitive = true
	description = "The secret key."
	final = true
	account_id = "%[1]s"
}
`, defaultAccount)

var testAccScalrVariablesDataSourceConfig = testAccScalrVariablesDataSourceInitConfig + fmt.Sprintf(`
data "scalr_variables" "shell" {
  category = "shell"
}

data "scalr_variables" "host" {
	keys = ["host"]
}

data "scalr_variables" "workspace" {
	workspace_ids=[scalr_workspace.test.id]
}

data "scalr_variables" "workspace_and_null" {
	workspace_ids=[scalr_workspace.test.id, "null"]
}

data "scalr_variables" "account" {
	account_id = "%[1]s"
}

`, defaultAccount)
