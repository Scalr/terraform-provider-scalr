package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/scalr/go-scalr/v2/scalr/schemas"
)

func TestAccScalrModuleNamespace_basic(t *testing.T) {
	namespaceName := acctest.RandomWithPrefix("test-namespace")
	namespace := &schemas.ModuleNamespace{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceBasic(namespaceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleNamespaceExists("scalr_module_namespace.test", namespace),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "name", namespaceName),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "is_shared", "false"),
				),
			},
		},
	})
}

func TestAccScalrModuleNamespace_import(t *testing.T) {
	namespaceName := acctest.RandomWithPrefix("test-namespace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceBasic(namespaceName),
			},
			{
				ResourceName:      "scalr_module_namespace.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalrModuleNamespace_update(t *testing.T) {
	namespaceName := acctest.RandomWithPrefix("test-namespace")
	namespaceNameUpdated := acctest.RandomWithPrefix("test-namespace-updated")
	namespace := &schemas.ModuleNamespace{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceBasic(namespaceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleNamespaceExists("scalr_module_namespace.test", namespace),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "name", namespaceName),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "is_shared", "false"),
				),
			},
			{
				Config: testAccScalrModuleNamespaceWithShared(namespaceNameUpdated, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleNamespaceExists("scalr_module_namespace.test", namespace),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "name", namespaceNameUpdated),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "is_shared", "true"),
				),
			},
		},
	})
}

func TestAccScalrModuleNamespace_emptyCollections(t *testing.T) {
	namespaceName := acctest.RandomWithPrefix("test-namespace")
	namespace := &schemas.ModuleNamespace{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceEmptyCollections(namespaceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrModuleNamespaceExists("scalr_module_namespace.test", namespace),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "name", namespaceName),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "environments.#", "0"),
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "owners.#", "0"),
				),
			},
		},
	})
}

func TestAccScalrModuleNamespace_environmentsUnmanaged(t *testing.T) {
	namespaceName := acctest.RandomWithPrefix("test-namespace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrModuleNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrModuleNamespaceWithEnvironment(namespaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "environments.#", "1"),
				),
			},
			{
				// Dropping the attribute means the links are no longer managed here,
				// so the namespace keeps them and the plan is empty.
				Config: testAccScalrModuleNamespaceWithEnvironmentUnmanaged(namespaceName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "environments.#", "1"),
				),
			},
			{
				// An explicit empty set still removes every link.
				Config: testAccScalrModuleNamespaceWithEnvironmentEmptied(namespaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_module_namespace.test", "environments.#", "0"),
				),
			},
		},
	})
}

func TestAccScalrModuleNamespace_invalidName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrModuleNamespaceInvalidName(),
				ExpectError: regexp.MustCompile("must only contain letters, numbers, dashes, and underscores"),
			},
		},
	})
}

func testAccScalrModuleNamespaceBasic(name string) string {
	return fmt.Sprintf(`
resource "scalr_module_namespace" "test" {
  name = "%s"
}`, name)
}

func testAccScalrModuleNamespaceWithShared(name string, isShared bool) string {
	return fmt.Sprintf(`
resource "scalr_module_namespace" "test" {
  name      = "%s"
  is_shared = %t
}`, name, isShared)
}

func testAccScalrModuleNamespaceEmptyCollections(name string) string {
	return fmt.Sprintf(`
resource "scalr_module_namespace" "test" {
  name         = "%s"
  environments = []
  owners       = []
}`, name)
}

func testAccScalrModuleNamespaceWithEnvironment(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s-env"
  account_id = "%[2]s"
}

resource "scalr_module_namespace" "test" {
  name         = "%[1]s"
  environments = [scalr_environment.test.id]
}`, name, defaultAccount)
}

func testAccScalrModuleNamespaceWithEnvironmentUnmanaged(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s-env"
  account_id = "%[2]s"
}

resource "scalr_module_namespace" "test" {
  name = "%[1]s"
}`, name, defaultAccount)
}

func testAccScalrModuleNamespaceWithEnvironmentEmptied(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s-env"
  account_id = "%[2]s"
}

resource "scalr_module_namespace" "test" {
  name         = "%[1]s"
  environments = []
}`, name, defaultAccount)
}

func testAccScalrModuleNamespaceInvalidName() string {
	return `
resource "scalr_module_namespace" "test" {
  name = "invalid-name@#$%"
}`
}

func testAccCheckScalrModuleNamespaceExists(resId string, namespace *schemas.ModuleNamespace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		n, err := scalrClient.ModuleNamespace.GetModuleNamespace(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*namespace = *n

		return nil
	}
}

func testAccCheckScalrModuleNamespaceDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_module_namespace" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.ModuleNamespace.GetModuleNamespace(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Module Namespace %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
