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
				Config: testAccScalrVariable_basic,
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
		},
	})
}

func TestAccScalrVariable_update(t *testing.T) {
	variable := &scalr.Variable{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVariable_basic,
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
				Config: testAccScalrVariable_update,
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
				Config: testAccScalrVariable_basic_nonsensitive,
			},

			{
				ResourceName:        "scalr_variable.foobar",
				ImportState:         true,
				ImportStateIdPrefix: "existing-env/existing-ws/",
				ImportStateVerify:   true,
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

const testAccScalrVariable_basic = `
resource "scalr_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  category     = "env"
  workspace_id = "existing-ws"
  sensitive    = true
}`

const testAccScalrVariable_basic_nonsensitive = `
resource "scalr_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  category     = "env"
  workspace_id = "existing-ws"
  sensitive    = false
}`

const testAccScalrVariable_update = `
resource "scalr_variable" "foobar" {
  key          = "key_updated"
  value        = "value_updated"
  category     = "terraform"
  hcl          = true
  sensitive    = false
  workspace_id = "existing-ws"
}`
