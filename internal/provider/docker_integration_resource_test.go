package provider

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr/v2/scalr/client"
)

// dockerRegistryTestCreds returns the registry URL, username and password used by Docker
// integration tests. Scalr validates the connection, so they must point to a real registry.
func dockerRegistryTestCreds(t *testing.T) (string, string, string) {
	registryURL := os.Getenv("TEST_DOCKER_REGISTRY_URL")
	username := os.Getenv("TEST_DOCKER_REGISTRY_USERNAME")
	password := os.Getenv("TEST_DOCKER_REGISTRY_PASSWORD")
	if registryURL == "" || username == "" || password == "" {
		t.Skip("Please set TEST_DOCKER_REGISTRY_URL, TEST_DOCKER_REGISTRY_USERNAME and TEST_DOCKER_REGISTRY_PASSWORD to run this test")
	}
	return registryURL, username, password
}

func TestNormalizeRegistryURL(t *testing.T) {
	cases := map[string]string{
		"ghcr.io":                   "https://ghcr.io",
		"https://ghcr.io/":          "https://ghcr.io",
		"GHCR.io:443":               "https://ghcr.io",
		" https://Registry.io:5000": "https://registry.io:5000",
		"http://localhost:80":       "http://localhost",
		"http://localhost:8080":     "http://localhost:8080",
		"https://[::1]:5000":        "https://[::1]:5000",
	}
	for in, want := range cases {
		if got := normalizeRegistryURL(in); got != want {
			t.Errorf("normalizeRegistryURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAccScalrDockerIntegrationResource_basic(t *testing.T) {
	registryURL, username, password := dockerRegistryTestCreds(t)
	name := acctest.RandomWithPrefix("test-docker")
	// Use the host-only spelling to check the normalized API value doesn't show up as drift.
	hostOnly := strings.TrimPrefix(normalizeRegistryURL(registryURL), "https://")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrDockerIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrDockerIntegrationConfig(name, hostOnly, username, password, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrDockerIntegrationExists("scalr_docker_integration.test"),
					resource.TestCheckResourceAttrSet("scalr_docker_integration.test", "id"),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "name", name),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "registry_url", hostOnly),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "username", username),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "password", password),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "export_credentials", "false"),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "status", "active"),
				),
			},
			{
				Config: testAccScalrDockerIntegrationConfig(name, hostOnly, username, password, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccScalrDockerIntegrationResource_update(t *testing.T) {
	registryURL, username, password := dockerRegistryTestCreds(t)
	name := acctest.RandomWithPrefix("test-docker")
	newName := acctest.RandomWithPrefix("test-docker")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrDockerIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrDockerIntegrationConfig(name, registryURL, username, password, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "name", name),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "export_credentials", "false"),
				),
			},
			{
				Config: testAccScalrDockerIntegrationConfig(newName, registryURL, username, password, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrDockerIntegrationExists("scalr_docker_integration.test"),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "name", newName),
					resource.TestCheckResourceAttr("scalr_docker_integration.test", "export_credentials", "true"),
				),
			},
			{
				Config:      testAccScalrDockerIntegrationConfig(newName, registryURL, username, "wrong-password", true),
				ExpectError: regexp.MustCompile("Error updating Docker integration"),
			},
		},
	})
}

func TestAccScalrDockerIntegrationResource_import(t *testing.T) {
	registryURL, username, password := dockerRegistryTestCreds(t)
	name := acctest.RandomWithPrefix("test-docker")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrDockerIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrDockerIntegrationConfig(name, normalizeRegistryURL(registryURL), username, password, false),
			},
			{
				ResourceName:            "scalr_docker_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccCheckScalrDockerIntegrationExists(resId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := createScalrClientV2()

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.DockerIntegration.GetDockerIntegration(ctx, rs.Primary.ID, nil)
		return err
	}
}

func testAccCheckScalrDockerIntegrationDestroy(s *terraform.State) error {
	scalrClient := createScalrClientV2()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_docker_integration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.DockerIntegration.GetDockerIntegration(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Docker integration %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, client.ErrNotFound) {
			return err
		}
	}

	return nil
}

func testAccScalrDockerIntegrationConfig(name, registryURL, username, password string, exportCredentials bool) string {
	return fmt.Sprintf(`
resource "scalr_docker_integration" "test" {
  name               = "%s"
  registry_url       = "%s"
  username           = "%s"
  password           = "%s"
  export_credentials = %t
}`, name, registryURL, username, password, exportCredentials)
}
