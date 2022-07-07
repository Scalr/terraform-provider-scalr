package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	scalr "github.com/scalr/go-scalr"
)

const baseForUpdate = `
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%[2]s"
}
resource scalr_workspace test {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

`

func TestAccScalrVariable_basic(t *testing.T) {
	variable := &scalr.Variable{}
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableOnGlobalScope(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists("scalr_variable.test", variable),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "key", fmt.Sprintf("var_on_global_%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "value", "test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "category", "shell"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "description", "Test on global scope"),
				),
			},

			// Test creation of sensitive variable
			{
				PreConfig: func() { rInt++ },
				Config:    testAccScalrVariableOnGlobalScopeSensitive(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists("scalr_variable.test", variable),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "key", fmt.Sprintf("var_on_global_%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "value", "test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "category", "shell"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "sensitive", "true"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "description", "Test on global scope sensitive"),
				),
			},
		},
	})
}

func TestAccScalrVariable_defaults(t *testing.T) {
	rInt := GetRandomInteger()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableOnGlobalScope(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "force", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "final", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "description", "Test on global scope"),
				),
			},
		},
	})
}
func TestAccScalrVariable_scopes(t *testing.T) {
	rInt := GetRandomInteger()
	variable := &scalr.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableOnAllScopes(rInt),
				Check:  testAccCheckScalrVariableOnScopes(variable),
			},
		},
	})
}

func TestAccScalrVariable_notTerraformOnMultiscope(t *testing.T) {
	rInt := GetRandomInteger()
	r := regexp.MustCompile("Attribute 'workspace_id' is required for variable with category 'terraform'.")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrVariableNotTerraformOnMultiscope(rInt),
				ExpectError: r,
			},
		},
	})
}

func TestAccScalrVariable_update(t *testing.T) {
	rInt := GetRandomInteger()
	variable := &scalr.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableOnWorkspaceScope(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists(
						"scalr_variable.test", variable),
					testAccCheckScalrVariableAttributes(variable, rInt),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "key", fmt.Sprintf("var_on_ws_%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "value", "test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "category", "shell"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "force", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "final", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "description", "Test update"),
				),
			},

			{
				Config: testAccScalrVariableOnWorkspaceScopeUpdateValue(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists(
						"scalr_variable.test", variable),
					testAccCheckScalrVariableAttributesUpdate(variable, rInt),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "key", fmt.Sprintf("var_on_ws_updated_%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "value", "updated"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "category", "terraform"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "hcl", "true"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "force", "true"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "final", "true"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "description", "updated"),
				),
			},

			// Test change key attribute for sensitive variable
			{
				Config: testAccScalrVariableOnWorkspaceScopeUpdateSensitivity(rInt),
			},

			{
				Config:      testAccScalrVariableOnWorkspaceScopeUpdateSensitivity(rInt + 1),
				ExpectError: regexp.MustCompile("Error changing 'key' attribute for variable var-[a-z0-9]+: immutable for sensitive variable"),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccScalrVariable_import(t *testing.T) {
	rInt := GetRandomInteger()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableOnWorkspaceScope(rInt),
			},
			{
				ResourceName:      "scalr_variable.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func variableFromState(s *terraform.State, n string, v *scalr.Variable) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	rs, ok := s.RootModule().Resources[n]
	if !ok {
		return fmt.Errorf("Not found: %s", n)
	}

	if rs.Primary.ID == "" {
		return fmt.Errorf("No instance ID is set")
	}

	variable, err := scalrClient.Variables.Read(ctx, rs.Primary.ID)
	if err != nil {
		return err
	}
	*v = *variable
	return nil
}

func testAccCheckScalrVariableExists(
	n string, v *scalr.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return variableFromState(s, n, v)

	}
}

func testAccCheckScalrVariableOnScopes(v *scalr.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Check on global scope
		err := variableFromState(s, "scalr_variable.on_global", v)
		if err != nil {
			return err
		}
		if v.Account != nil || v.Environment != nil || v.Workspace != nil {
			return fmt.Errorf("Variable %s not on global scope.", v.ID)
		}
		// Check on account scope
		err = variableFromState(s, "scalr_variable.on_account", v)
		if err != nil {
			return err
		}
		if v.Account == nil || v.Environment != nil || v.Workspace != nil {
			return fmt.Errorf("Variable %s not on account scope.", v.ID)
		}
		// Check on environment scope
		err = variableFromState(s, "scalr_variable.on_environment", v)
		if err != nil {
			return err
		}
		if v.Account == nil || v.Environment == nil || v.Workspace != nil {
			return fmt.Errorf("Variable %s not on environment scope.", v.ID)
		}
		// Check on workspace scope
		err = variableFromState(s, "scalr_variable.on_workspace", v)
		if err != nil {
			return err
		}
		if v.Account == nil || v.Environment == nil || v.Workspace == nil {
			return fmt.Errorf("Variable %s not on workspace scope.", v.ID)
		}
		return nil
	}
}

func testAccCheckScalrVariableAttributes(
	variable *scalr.Variable, rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != fmt.Sprintf("var_on_ws_%d", rInt) {
			return fmt.Errorf("Bad key: %s != %s", variable.Key, variable.Key)
		}

		if variable.Value != "test" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Category != scalr.CategoryShell {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != false {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Final != false {
			return fmt.Errorf("Bad final: %t", variable.Final)
		}

		if variable.Sensitive != false {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
		}
		return nil
	}
}

func testAccCheckScalrVariableAttributesUpdate(
	variable *scalr.Variable, rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != fmt.Sprintf("var_on_ws_updated_%d", rInt) {
			return fmt.Errorf("Bad key: %s != %s", variable.Key, variable.Key)
		}

		if variable.Value != "updated" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Category != scalr.CategoryTerraform {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != true {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Final != true {
			return fmt.Errorf("Bad final: %t", variable.Final)
		}

		return nil
	}
}

func testAccCheckScalrVariableDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_variable" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Variables.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Variable %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrVariableOnGlobalScope(rInt int) string {
	return fmt.Sprintf(`
resource scalr_variable test {
  key          = "var_on_global_%d"
  value        = "test"
  category     = "shell"
  description  = "Test on global scope"
}`, rInt)
}

func testAccScalrVariableOnGlobalScopeSensitive(rInt int) string {
	return fmt.Sprintf(`
resource scalr_variable test {
  key          = "var_on_global_%d"
  value        = "test"
  category     = "shell"
  sensitive    = true
  description  = "Test on global scope sensitive"
}`, rInt)
}

func testAccScalrVariableOnWorkspaceScope(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%[2]s"
}

resource scalr_workspace test {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_variable test {
  key            = "var_on_ws_%[1]d"
  value          = "test"
  category       = "shell"
  workspace_id   = scalr_workspace.test.id
  description    = "Test update"
}`, rInt, defaultAccount)
}

func testAccScalrVariableNotTerraformOnMultiscope(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%[2]s"
}

resource scalr_workspace test {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_variable test {
  key            = "var_on_ws_%[1]d"
  value          = "test"
  category       = "terraform"
  account_id     = "%[2]s"
  environment_id = scalr_environment.test.id
  description  = "Test not terraform variable"
}`, rInt, defaultAccount)
}

func testAccScalrVariableOnAllScopes(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%[2]s"
}

resource scalr_workspace test {
  name           = "test-ws-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_variable on_global {
  key          = "var_on_global_%[1]d"
  value        = "test"
  category     = "shell"
}

resource scalr_variable on_account {
  key          = "var_on_acc_%[1]d"
  value        = "test"
  category     = "shell"
  account_id   = "%[2]s"
}

resource scalr_variable on_environment {
  key            = "var_on_env_%[1]d"
  value          = "test"
  category       = "shell"
  account_id     = "%[2]s"
  environment_id = scalr_environment.test.id
}

resource scalr_variable on_workspace {
  key            = "var_on_ws_%[1]d"
  value          = "test"
  category       = "shell"
  account_id     = "%[2]s"
  environment_id = scalr_environment.test.id
  workspace_id   = scalr_workspace.test.id
}`, rInt, defaultAccount)
}

func testAccScalrVariableOnWorkspaceScopeUpdateValue(rInt int) string {
	return fmt.Sprintf(baseForUpdate+`
resource scalr_variable test {
  key            = "var_on_ws_updated_%[1]d"
  value          = "updated"
  category       = "terraform"
  hcl            = true
  force          = true
  final          = true
  account_id     = "%[2]s"
  environment_id = scalr_environment.test.id
  workspace_id   = scalr_workspace.test.id
  description    = "updated"
}`, rInt, defaultAccount)
}

func testAccScalrVariableOnWorkspaceScopeUpdateSensitivity(rInt int) string {
	return fmt.Sprintf(baseForUpdate+`
resource scalr_variable test {
  key            = "var_on_ws_updated_%[1]d"
  value          = "updated"
  category       = "terraform"
  hcl            = true
  force          = true
  final          = true
  sensitive      = true
  account_id     = "%[2]s"
  environment_id = scalr_environment.test.id
  workspace_id   = scalr_workspace.test.id
}`, rInt, defaultAccount)
}
