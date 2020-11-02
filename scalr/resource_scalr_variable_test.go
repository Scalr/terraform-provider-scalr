package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccScalrVariable_basic(t *testing.T) {
	variable := &scalr.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists("scalr_variable.foobar", variable),
					testAccCheckScalrVariableAttributes(variable),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "sensitive", "true"),
				),
			},
		},
	})
}

// func GetResourceIDfromState(prefix *string, resourceDeclaration string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[resourceDeclaration]
// 		if !ok {
// 			return fmt.Errorf("Not found: %s", resourceDeclaration)
// 		}
// 		if rs.Primary.ID == "" {
// 			return fmt.Errorf("No instance ID is set")
// 		}
// 		*prefix = fmt.Sprintf("%s/test-ws/", rs.Primary.ID)
// 		println("I GOT: %s", *prefix)
// 		return nil
// 	}
// }

// func CheckValue(resourceID *string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		println("I GOT: %s", *resourceID)
// 		return nil
// 	}
// }

func TestAccScalrVariable_update(t *testing.T) {
	variable := &scalr.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists(
						"scalr_variable.foobar", variable),
					testAccCheckScalrVariableAttributes(variable),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "sensitive", "true"),
				),
			},

			{
				Config: testAccScalrVariableUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVariableExists(
						"scalr_variable.foobar", variable),
					testAccCheckScalrVariableAttributesUpdate(variable),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "key", "key_updated"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "value", "value_updated"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "category", "terraform"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "hcl", "true"),
					resource.TestCheckResourceAttr(
						"scalr_variable.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func TestAccScalrVariable_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariableBasicNonsensitive,
			},
			{
				ResourceName: "scalr_variable.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					resources := s.RootModule().Resources
					env := resources["scalr_environment.test"]
					variable := resources["scalr_variable.test"]
					return fmt.Sprintf("%s/test-ws/%s", env.Primary.ID, variable.Primary.ID), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrVariableExists(
	n string, variable *scalr.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		v, err := scalrClient.Variables.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*variable = *v

		return nil
	}
}

func testAccCheckScalrVariableAttributes(
	variable *scalr.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_test" {
			return fmt.Errorf("Bad key: %s", variable.Key)
		}

		if variable.Value != "" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Category != scalr.CategoryEnv {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != false {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Sensitive != true {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
		}

		return nil
	}
}

func testAccCheckScalrVariableAttributesUpdate(
	variable *scalr.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_updated" {
			return fmt.Errorf("Bad key: %s", variable.Key)
		}

		if variable.Value != "value_updated" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Category != scalr.CategoryTerraform {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != true {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Sensitive != false {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
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

const testAccScalrVariableCommonConfig = `
resource scalr_environment test {
  name       = "test-env"
  account_id = "acc-svrcncgh453bi8g"
}
  
resource scalr_workspace test {
  name           = "test-ws"
  environment_id = scalr_environment.test.id
}`

const testAccScalrVariableBasic = testAccScalrVariableCommonConfig + `
resource scalr_variable test {
  key          = "key_test"
  value        = "value_test"
  category     = "env"
  workspace_id = scalr_workspace.test.id
  sensitive    = true
}`

const testAccScalrVariableBasicNonsensitive = testAccScalrVariableCommonConfig + `
resource scalr_variable test {
  key          = "key_test"
  value        = "value_test"
  category     = "env"
  workspace_id = scalr_workspace.test.id 
  sensitive    = false
}`

const testAccScalrVariableUpdate = testAccScalrVariableCommonConfig + `
resource scalr_variable test {
  key          = "key_updated"
  value        = "value_updated"
  category     = "terraform"
  hcl          = true
  sensitive    = false
  workspace_id = scalr_workspace.test.id
}`
