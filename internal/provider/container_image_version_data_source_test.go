package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrContainerImageVersionDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_container_image_version test { container_image_id = "" }`,
				ExpectError: regexp.MustCompile("Attribute container_image_id must not be empty"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrContainerImageVersionDataSourceLatestConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_container_image_version.latest", "id",
						"data.scalr_container_image.runner", "latest_version_id",
					),
					resource.TestCheckResourceAttrSet("data.scalr_container_image_version.latest", "version"),
					resource.TestCheckResourceAttr("data.scalr_container_image_version.latest", "latest", "true"),
					resource.TestCheckResourceAttr("data.scalr_container_image_version.latest", "available", "true"),
					resource.TestCheckResourceAttrSet("data.scalr_container_image_version.latest", "created_at"),
				),
			},
			{
				Config: testAccScalrContainerImageVersionDataSourceByVersionConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scalr_container_image_version.by_version", "id",
						"data.scalr_container_image_version.latest", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scalr_container_image_version.by_version", "version",
						"data.scalr_container_image_version.latest", "version",
					),
				),
			},
			{
				Config:      testAccScalrContainerImageVersionDataSourceMissingConfig,
				ExpectError: regexp.MustCompile(`Could not find version '0.0.0-does-not-exist'`),
			},
		},
	})
}

var testAccScalrContainerImageVersionDataSourceLatestConfig = `
data scalr_container_image runner {
  name       = "scalr/runner"
  visibility = "system"
}

data scalr_container_image_version latest {
  container_image_id = data.scalr_container_image.runner.id
}`

var testAccScalrContainerImageVersionDataSourceByVersionConfig = testAccScalrContainerImageVersionDataSourceLatestConfig + `

data scalr_container_image_version by_version {
  container_image_id = data.scalr_container_image.runner.id
  version            = data.scalr_container_image_version.latest.version
}`

var testAccScalrContainerImageVersionDataSourceMissingConfig = testAccScalrContainerImageVersionDataSourceLatestConfig + `

data scalr_container_image_version missing {
  container_image_id = data.scalr_container_image.runner.id
  version            = "0.0.0-does-not-exist"
}`
