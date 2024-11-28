package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrVariableDataSource(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_variable test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,key` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_variable test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_variable test {key = ""}`,
				ExpectError: regexp.MustCompile("expected \"key\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrVariableDataSourceInitConfig, // depends_on works improperly with data sources
			},
			{
				Config: testAccScalrVariableDataSourceByIDConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_variable.secret", "scalr_variable.secret"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "hcl", "false"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "sensitive", "true"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "description", "The secret key."),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "final", "true"),
				),
			},
			{
				Config: testAccScalrVariableDataSourceByIDAndKeyConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_variable.secret", "scalr_variable.secret"),
				),
			},
			{
				Config: testAccScalrVariableDataSourceByKeyConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEqualID("data.scalr_variable.workspace_hostname", "scalr_variable.workspace_hostname"),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "hcl", "false"),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "sensitive", "false"),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "description", "The hostname of scalr workspace."),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "final", "false"),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "account_id", defaultAccount),
					testAccCheckEqualID("data.scalr_variable.secret", "scalr_variable.secret"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "hcl", "false"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "sensitive", "true"),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "description", "The secret key."),
					resource.TestCheckResourceAttr("data.scalr_variable.secret", "final", "true"),
					resource.TestCheckResourceAttrSet("data.scalr_variable.workspace_hostname", "updated_at"),
					resource.TestCheckResourceAttrSet("data.scalr_variable.workspace_hostname", "updated_by_email"),
					resource.TestCheckResourceAttr("data.scalr_variable.workspace_hostname", "updated_by.#", "1"),
					resource.TestCheckResourceAttrSet("data.scalr_variable.workspace_hostname", "updated_by.0.username"),
					resource.TestCheckResourceAttrSet("data.scalr_variable.workspace_hostname", "updated_by.0.email"),
					resource.TestCheckResourceAttrSet("data.scalr_variable.workspace_hostname", "updated_by.0.full_name"),
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
  key          = "hostname"
  value        = "workspace.scalr.com"
  category     = "shell"
  hcl          = false
  sensitive    = false
  description  = "The hostname of scalr workspace."
  final        = false
  workspace_id = scalr_workspace.test.id
}

resource "scalr_variable" "hostname" {
  key         = "hostname"
  value       = "scalr.com"
  category    = "shell"
  hcl         = false
  sensitive   = false
  description = "The hostname of scalr."
  final       = false
  account_id  = "%[1]s"
}

resource "scalr_variable" "secret" {
  key         = "secret"
  value       = "secret-key"
  category    = "shell"
  hcl         = false
  sensitive   = true
  description = "The secret key."
  final       = true
  account_id  = "%[1]s"
}
`, defaultAccount)

var testAccScalrVariableDataSourceByIDConfig = testAccScalrVariableDataSourceInitConfig + fmt.Sprintf(`
data "scalr_variable" "secret" {
  id         = scalr_variable.secret.id
  account_id = "%s"
}
`, defaultAccount)

var testAccScalrVariableDataSourceByKeyConfig = testAccScalrVariableDataSourceInitConfig + fmt.Sprintf(`
data "scalr_variable" "secret" {
  key        = "secret"
  category   = "shell"
  account_id = "%[1]s"
}

data "scalr_variable" "workspace_hostname" {
  key = "hostname"
  category = "shell"
  account_id = "%[1]s"
  workspace_id = scalr_workspace.test.id
}`, defaultAccount)

var testAccScalrVariableDataSourceByIDAndKeyConfig = testAccScalrVariableDataSourceInitConfig + fmt.Sprintf(`
data "scalr_variable" "secret" {
  id         = scalr_variable.secret.id
  key        = "secret"
  account_id = "%s"
}
`, defaultAccount)
