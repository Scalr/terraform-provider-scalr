package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccScalrVarSetResource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-var-set")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrVarSetDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrVarSetBasic(name),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrVarSetExists("scalr_var_set.test"),
						resource.TestCheckResourceAttr("scalr_var_set.test", "name", name),
						resource.TestCheckResourceAttr("scalr_var_set.test", "account_id", defaultAccount),
						resource.TestCheckResourceAttrSet("scalr_var_set.test", "updated_at"),
						resource.TestCheckResourceAttrSet("scalr_var_set.test", "updated_by_email"),
						resource.TestCheckResourceAttr("scalr_var_set.test", "environments.#", "0"),
						resource.TestCheckResourceAttr("scalr_var_set.test", "owners.#", "0"),
					),
				},
				{
					Config: testAccScalrVarSetBasic(name),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectEmptyPlan(),
						},
					},
				},
			},
		},
	)
}

func TestAccScalrVarSetResource_update(t *testing.T) {
	name := acctest.RandomWithPrefix("test-var-set")
	newName := acctest.RandomWithPrefix("test-var-set")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrVarSetDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrVarSetBasic(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("scalr_var_set.test", "name", name),
					),
				},
				{
					Config: testAccScalrVarSetWithDescription(newName, "Updated description"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("scalr_var_set.test", "name", newName),
						resource.TestCheckResourceAttr("scalr_var_set.test", "description", "Updated description"),
					),
				},
			},
		},
	)
}

func TestAccScalrVarSetResource_environments(t *testing.T) {
	name := acctest.RandomWithPrefix("test-var-set")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrVarSetDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrVarSetSharedWithAll(name),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckScalrVarSetExists("scalr_var_set.test"),
						resource.TestCheckResourceAttr("scalr_var_set.test", "environments.#", "1"),
						resource.TestCheckTypeSetElemAttr("scalr_var_set.test", "environments.*", "*"),
					),
				},
				{
					Config: testAccScalrVarSetBasic(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("scalr_var_set.test", "environments.#", "0"),
					),
				},
			},
		},
	)
}

func TestAccScalrVarSetResource_validation(t *testing.T) {
	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			Steps: []resource.TestStep{
				{
					Config:      testAccScalrVarSetWhitespaceName(),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("Attribute name must not be empty"),
				},
			},
		},
	)
}

func TestAccScalrVarSetResource_import(t *testing.T) {
	name := acctest.RandomWithPrefix("test-var-set")

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: protoV5ProviderFactories(t),
			CheckDestroy:             testAccCheckScalrVarSetDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccScalrVarSetWithDescription(name, "import test"),
				},
				{
					ResourceName:      "scalr_var_set.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		},
	)
}

func testAccCheckScalrVarSetExists(resID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resID]
		if !ok {
			return fmt.Errorf("not found: %s", resID)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := scalrClient.VariableSet.GetVarSet(ctx, rs.Primary.ID, nil)
		return err
	}
}

func testAccCheckScalrVarSetDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_var_set" {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}
		_, err := scalrClient.VariableSet.GetVarSet(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("var set %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrVarSetBasic(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_var_set" "test" {
  name = "%s"
}`, name,
	)
}

func testAccScalrVarSetWithDescription(name, description string) string {
	return fmt.Sprintf(
		`
resource "scalr_var_set" "test" {
  name        = "%s"
  description = "%s"
}`, name, description,
	)
}

func testAccScalrVarSetSharedWithAll(name string) string {
	return fmt.Sprintf(
		`
resource "scalr_var_set" "test" {
  name         = "%s"
  environments = ["*"]
}`, name,
	)
}

func testAccScalrVarSetWhitespaceName() string {
	return `
resource "scalr_var_set" "test" {
  name = "   "
}`
}
