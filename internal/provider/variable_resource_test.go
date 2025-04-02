package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableOnAccountScopeImplicit(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists("scalr_variable.test", variable),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "key", fmt.Sprintf("var_on_account_%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "value", "test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "category", "shell"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "description", "Test on account scope"),
				),
			},

			// Test creation of sensitive variable
			{
				PreConfig: func() { rInt++ },
				Config:    testAccScalrVariableOnAccountScopeSensitive(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists("scalr_variable.test", variable),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "key", fmt.Sprintf("var_on_account_%d", rInt)),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "value", "test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "category", "shell"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "sensitive", "true"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "description", "Test on account scope sensitive"),
					resource.TestCheckResourceAttrSet(
						"scalr_variable.test", "updated_at"),
					resource.TestCheckResourceAttrSet(
						"scalr_variable.test", "updated_by_email"),
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "updated_by.#", "1"),
					resource.TestCheckResourceAttrSet(
						"scalr_variable.test", "updated_by.0.username"),
					resource.TestCheckResourceAttrSet(
						"scalr_variable.test", "updated_by.0.email"),
					resource.TestCheckResourceAttrSet(
						"scalr_variable.test", "updated_by.0.full_name"),
				),
			},
		},
	})
}

func TestAccScalrVariable_defaults(t *testing.T) {
	rInt := GetRandomInteger()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableOnAccountScopeImplicit(rInt),
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
						"scalr_variable.test", "description", "Test on account scope"),
				),
			},
		},
	})
}
func TestAccScalrVariable_scopes(t *testing.T) {
	rInt := GetRandomInteger()
	variable := &scalr.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableOnAllScopes(rInt),
				Check:  testAccCheckScalrVariableOnScopes(variable),
			},
		},
	})
}

func TestAccScalrVariable_update(t *testing.T) {
	rInt := GetRandomInteger()
	variable := &scalr.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrVariableDestroy,
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
					resource.TestCheckResourceAttr(
						"scalr_variable.test", "account_id", defaultAccount),
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
		},
	})
}

func TestAccScalrVariable_import(t *testing.T) {
	rInt := GetRandomInteger()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrVariableDestroy,
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

func TestAccScalrVariable_UpgradeFromSDK(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"scalr": {
						Source:            "registry.scalr.io/scalr/scalr",
						VersionConstraint: "<=2.6.0",
					},
				},
				Config: testAccScalrVariableOnAccountScopeImplicit(rInt),
				Check:  resource.TestCheckResourceAttrSet("scalr_variable.test", "id"),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(t),
				Config:                   testAccScalrVariableOnAccountScopeImplicit(rInt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func variableFromState(s *terraform.State, n string, v *scalr.Variable) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

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
		// Check on account (implicit) scope
		err := variableFromState(s, "scalr_variable.on_account_implicit", v)
		if err != nil {
			return err
		}
		if v.Account == nil || v.Environment != nil || v.Workspace != nil {
			return fmt.Errorf("Variable %s not on account scope.", v.ID)
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

		if variable.UpdatedBy == nil || variable.UpdatedBy.Email != variable.UpdatedByEmail {
			return fmt.Errorf("Bad updated by")
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
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

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

func testAccScalrVariableOnAccountScopeImplicit(rInt int) string {
	return fmt.Sprintf(`
resource scalr_variable test {
  key          = "var_on_account_%d"
  value        = "test"
  category     = "shell"
  description  = "Test on account scope"
}`, rInt)
}

func testAccScalrVariableOnAccountScopeSensitive(rInt int) string {
	return fmt.Sprintf(`
resource scalr_variable test {
  key          = "var_on_account_%d"
  value        = "test"
  category     = "shell"
  sensitive    = true
  description  = "Test on account scope sensitive"
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

resource scalr_variable on_account_implicit {
  key          = "var_on_acc_impl_%[1]d"
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
