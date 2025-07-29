package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrModuleNamespace_basic(t *testing.T) {
	namespaceName := acctest.RandomWithPrefix("test-namespace")
	namespace := &scalr.ModuleNamespace{}

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
	namespace := &scalr.ModuleNamespace{}

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

func testAccScalrModuleNamespaceInvalidName() string {
	return `
resource "scalr_module_namespace" "test" {
  name = "invalid-name@#$%"
}`
}

func testAccCheckScalrModuleNamespaceExists(resId string, namespace *scalr.ModuleNamespace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		n, err := scalrClient.ModuleNamespaces.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*namespace = *n

		return nil
	}
}

func testAccCheckScalrModuleNamespaceDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_module_namespace" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.ModuleNamespaces.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Module Namespace %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
