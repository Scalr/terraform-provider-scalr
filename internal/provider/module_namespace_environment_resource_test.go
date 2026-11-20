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

func TestAccScalrModuleNamespaceEnvironment_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("test-namespace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceEnvironmentBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleNamespaceEnvironmentExists("scalr_module_namespace_environment.test"),
					resource.TestCheckResourceAttrPair(
						"scalr_module_namespace_environment.test", "module_namespace_id",
						"scalr_module_namespace.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"scalr_module_namespace_environment.test", "environment_id",
						"scalr_environment.test", "id",
					),
				),
			},
			{
				Config: testAccScalrModuleNamespaceEnvironmentBasic(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccScalrModuleNamespaceEnvironment_import(t *testing.T) {
	name := acctest.RandomWithPrefix("test-namespace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceEnvironmentBasic(name),
			},
			{
				ResourceName:      "scalr_module_namespace_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Each link is independent: removing one leaves the other in place, and neither
// is affected by the environments attribute of the namespace, which is omitted here.
func TestAccScalrModuleNamespaceEnvironment_independentLinks(t *testing.T) {
	name := acctest.RandomWithPrefix("test-namespace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceEnvironmentTwoLinks(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleNamespaceEnvironmentExists("scalr_module_namespace_environment.first"),
					testAccCheckScalrModuleNamespaceEnvironmentExists("scalr_module_namespace_environment.second"),
				),
			},
			{
				Config: testAccScalrModuleNamespaceEnvironmentOneLink(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleNamespaceEnvironmentExists("scalr_module_namespace_environment.first"),
				),
			},
		},
	})
}

func TestAccScalrModuleNamespaceEnvironment_sharedNamespace(t *testing.T) {
	name := acctest.RandomWithPrefix("test-namespace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrModuleNamespaceEnvironmentShared(name),
				ExpectError: regexp.MustCompile("shared module namespace"),
			},
		},
	})
}

func TestAccScalrModuleNamespaceEnvironment_invalidImportID(t *testing.T) {
	name := acctest.RandomWithPrefix("test-namespace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceEnvironmentBasic(name),
			},
			{
				ResourceName:  "scalr_module_namespace_environment.test",
				ImportState:   true,
				ImportStateId: "modns-xxxxxxxxxx",
				ExpectError:   regexp.MustCompile("Expected import ID in the format"),
			},
		},
	})
}

func testAccScalrModuleNamespaceEnvironmentBasic(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s-env"
  account_id = "%[2]s"
}

resource "scalr_module_namespace" "test" {
  name = "%[1]s"
}

resource "scalr_module_namespace_environment" "test" {
  module_namespace_id = scalr_module_namespace.test.id
  environment_id      = scalr_environment.test.id
}`, name, defaultAccount)
}

func testAccScalrModuleNamespaceEnvironmentTwoLinks(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "first" {
  name       = "%[1]s-env-1"
  account_id = "%[2]s"
}

resource "scalr_environment" "second" {
  name       = "%[1]s-env-2"
  account_id = "%[2]s"
}

resource "scalr_module_namespace" "test" {
  name = "%[1]s"
}

resource "scalr_module_namespace_environment" "first" {
  module_namespace_id = scalr_module_namespace.test.id
  environment_id      = scalr_environment.first.id
}

resource "scalr_module_namespace_environment" "second" {
  module_namespace_id = scalr_module_namespace.test.id
  environment_id      = scalr_environment.second.id
}`, name, defaultAccount)
}

func testAccScalrModuleNamespaceEnvironmentOneLink(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "first" {
  name       = "%[1]s-env-1"
  account_id = "%[2]s"
}

resource "scalr_environment" "second" {
  name       = "%[1]s-env-2"
  account_id = "%[2]s"
}

resource "scalr_module_namespace" "test" {
  name = "%[1]s"
}

resource "scalr_module_namespace_environment" "first" {
  module_namespace_id = scalr_module_namespace.test.id
  environment_id      = scalr_environment.first.id
}`, name, defaultAccount)
}

func testAccScalrModuleNamespaceEnvironmentShared(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s-env"
  account_id = "%[2]s"
}

resource "scalr_module_namespace" "test" {
  name      = "%[1]s"
  is_shared = true
}

resource "scalr_module_namespace_environment" "test" {
  module_namespace_id = scalr_module_namespace.test.id
  environment_id      = scalr_environment.test.id
}`, name, defaultAccount)
}

func testAccCheckScalrModuleNamespaceEnvironmentExists(resId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		namespaceID := rs.Primary.Attributes["module_namespace_id"]
		environmentID := rs.Primary.Attributes["environment_id"]

		namespace, err := scalrClient.ModuleNamespace.GetModuleNamespace(ctx, namespaceID)
		if err != nil {
			return err
		}

		for _, env := range namespace.Relationships.Environments {
			if env != nil && env.ID == environmentID {
				return nil
			}
		}

		return fmt.Errorf("Environment %s is not linked to module namespace %s", environmentID, namespaceID)
	}
}

func testAccCheckScalrModuleNamespaceEnvironmentDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_module_namespace_environment" {
			continue
		}

		namespaceID := rs.Primary.Attributes["module_namespace_id"]
		environmentID := rs.Primary.Attributes["environment_id"]

		namespace, err := scalrClient.ModuleNamespace.GetModuleNamespace(ctx, namespaceID)
		if err != nil {
			// The namespace is gone, so the link is gone with it.
			continue
		}

		for _, env := range namespace.Relationships.Environments {
			if env != nil && env.ID == environmentID {
				return fmt.Errorf(
					"Environment %s is still linked to module namespace %s",
					environmentID, namespaceID,
				)
			}
		}
	}

	return nil
}
