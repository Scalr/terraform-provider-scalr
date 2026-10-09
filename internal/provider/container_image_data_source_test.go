package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrContainerImageDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_container_image test {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_container_image test { name = "" }`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
				PlanOnly:    true,
			},
			{
				Config: `
data scalr_container_image test {
  name       = "scalr/runner"
  visibility = "private"
}`,
				ExpectError: regexp.MustCompile(`Attribute visibility value must be one of`),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_container_image test { name = "scalr/does-not-exist" }`,
				ExpectError: regexp.MustCompile(`Could not find container image with name 'scalr/does-not-exist'`),
			},
			{
				Config: testAccScalrContainerImageDataSourceByNameConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.scalr_container_image.test", "id", regexp.MustCompile(`^cimg-`)),
					resource.TestCheckResourceAttr("data.scalr_container_image.test", "name", "scalr/runner"),
					resource.TestCheckResourceAttr("data.scalr_container_image.test", "visibility", "system"),
					resource.TestCheckResourceAttrSet("data.scalr_container_image.test", "registry_url"),
					resource.TestCheckResourceAttrSet("data.scalr_container_image.test", "repository"),
					resource.TestCheckResourceAttrSet("data.scalr_container_image.test", "sync_status"),
					resource.TestMatchResourceAttr(
						"data.scalr_container_image.test", "latest_version_id", regexp.MustCompile(`^cimgv-`),
					),
				),
			},
			{
				Config: testAccScalrContainerImageDataSourceByIDConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_container_image.by_id", "id",
						"data.scalr_container_image.test", "id",
					),
					resource.TestCheckResourceAttr("data.scalr_container_image.by_id", "name", "scalr/runner"),
				),
			},
		},
	})
}

var testAccScalrContainerImageDataSourceByNameConfig = `
data scalr_container_image test {
  name       = "scalr/runner"
  visibility = "system"
}`

var testAccScalrContainerImageDataSourceByIDConfig = testAccScalrContainerImageDataSourceByNameConfig + `

data scalr_container_image by_id {
  id = data.scalr_container_image.test.id
}`
