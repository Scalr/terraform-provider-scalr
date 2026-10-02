package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrDockerIntegrationDataSource_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_docker_integration test {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
			},
			{
				Config:      `data scalr_docker_integration test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data scalr_docker_integration test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config:      `data scalr_docker_integration test {id = "int-nonexistent"}`,
				ExpectError: regexp.MustCompile("Could not find Docker integration with ID 'int-nonexistent'"),
			},
			{
				Config:      `data scalr_docker_integration test {name = "int-nonexistent"}`,
				ExpectError: regexp.MustCompile("Could not find Docker integration with name 'int-nonexistent'"),
			},
		},
	})
}

func TestAccScalrDockerIntegrationDataSource_basic(t *testing.T) {
	registryURL, username, password := dockerRegistryTestCreds(t)
	name := acctest.RandomWithPrefix("test-docker")
	res := testAccScalrDockerIntegrationConfig(name, registryURL, username, password, true)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: res + `
data "scalr_docker_integration" "test" {
  id = scalr_docker_integration.test.id
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_docker_integration.test", "id",
						"scalr_docker_integration.test", "id",
					),
					resource.TestCheckResourceAttr("data.scalr_docker_integration.test", "name", name),
					resource.TestCheckResourceAttr("data.scalr_docker_integration.test", "registry_url", normalizeRegistryURL(registryURL)),
					resource.TestCheckResourceAttr("data.scalr_docker_integration.test", "username", username),
					resource.TestCheckResourceAttr("data.scalr_docker_integration.test", "export_credentials", "true"),
					resource.TestCheckResourceAttr("data.scalr_docker_integration.test", "status", "active"),
				),
			},
			{
				Config: res + fmt.Sprintf(`
data "scalr_docker_integration" "test" {
  name = "%s"
}`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_docker_integration.test", "id",
						"scalr_docker_integration.test", "id",
					),
				),
			},
			{
				Config: res + `
data "scalr_docker_integration" "test" {
  id   = scalr_docker_integration.test.id
  name = "wrong-name"
}`,
				ExpectError: regexp.MustCompile(`Could not find Docker integration with ID '[^']+', name\s+'wrong-name'`),
			},
		},
	})
}
