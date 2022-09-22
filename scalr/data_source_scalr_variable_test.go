package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalrVariableDataSource(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config: testAccScalrVariableDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_variable.workspace_hostname", "scalr_variable.workspace_hostname"),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "hcl", "false"),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "sensitive", "false"),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "description", "The hostname of scalr workspace."),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "final", "false"),
					testAccCheckEqualID("data.scalr_variable.secret", "scalr_variable.secret"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "hcl", "false"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "sensitive", "true"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "description", "The secret key."),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "final", "true"),
				),
			},
			{
				Config: testAccScalrVariableDataSourceInitConfig,
			},
		},
	})
}

var testAccScalrVariableDataSourceInitConfig = fmt.Sprintf(`
resource scalr_environment test {
	name       = "test-env-variable-data"
	account_id = "%[1]s"
  }
  
resource scalr_workspace test {
name           = "test-ws-variable-data"
environment_id = scalr_environment.test.id
}
  
resource "scalr_variable" "workspace_hostname" {
	key = "hostname"
	value = "workspace.scalr.com"
	category = "shell"
	hcl = false
	sensitive = false
	description = "The hostname of scalr workspace."
	final = false
	workspace_id=scalr_workspace.test.id
}

resource "scalr_variable" "hostname" {
	key = "hostname"
	value = "scalr.com"
	category = "shell"
	hcl = false
	sensitive = false
	description = "The hostname of scalr."
	final = false
	account_id = "%[1]s"
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

var testAccScalrVariableDataSourceConfig = testAccScalrVariableDataSourceInitConfig + fmt.Sprintf(`
data "scalr_variable" "secret" {
  key = "secret"
  category = "shell"
  account_id = "%[1]s"
}

data "scalr_variable" "workspace_hostname" {
	key = "hostname"
	category = "shell"
	account_id = "%[1]s"
	workspace_id = scalr_workspace.test.id
}

`, defaultAccount)
