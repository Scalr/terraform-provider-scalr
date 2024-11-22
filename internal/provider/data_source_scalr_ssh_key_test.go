package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalrSSHKeyDataSource_basic(t *testing.T) {
	rInt := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_ssh_key test {}`,
				ExpectError: regexp.MustCompile("\"id\": one of `id,name` must be specified"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_ssh_key test {id = ""}`,
				ExpectError: regexp.MustCompile("expected \"id\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config:      `data scalr_ssh_key test {name = ""}`,
				ExpectError: regexp.MustCompile("expected \"name\" to not be an empty string or whitespace"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrSSHKeyDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_ssh_key.test", "name", fmt.Sprintf("ssh-key-test-%s", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_ssh_key.test", "environments.#"),
					resource.TestCheckResourceAttr(
						"data.scalr_ssh_key.test", "account_id", defaultAccount),
				),
			},
			{
				Config: testAccScalrSSHKeyDataSourceAccessByNameConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_ssh_key.test_by_name", "name", fmt.Sprintf("ssh-key-test-%s", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_ssh_key.test_by_name", "account_id", defaultAccount),
				),
			},
			{
				Config:      testAccScalrSSHKeyDataSourceNotFoundByNameConfig(),
				ExpectError: regexp.MustCompile("no SSH key found with name: ssh-key-foo-bar-baz"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrSSHKeyDataSourceMismatchIDNameConfig(rInt),
				ExpectError: regexp.MustCompile(
					fmt.Sprintf("SSH key name mismatch: the provided SSH key name 'ssh-key-test-%s' does not match the expected name 'incorrect-name'", rInt),
				),
				PlanOnly: true,
			},
		},
	})
}

func testAccScalrSSHKeyDataSourceConfig(rInt string) string {
	return fmt.Sprintf(`
resource "scalr_ssh_key" "test" {
  name         = "ssh-key-test-%s"
  private_key  = <<EOF
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBvMDyNaYtWK2TmJIfFhmPZeGxK0bWnNDhjlTZ+V6e4x
-----END PRIVATE KEY-----
EOF
  account_id   = "%s"
  environments = ["env-svrcnchebt61e30"]
}

data "scalr_ssh_key" "test" {
  id         = scalr_ssh_key.test.id
  account_id = scalr_ssh_key.test.account_id
}`, rInt, defaultAccount)
}

func testAccScalrSSHKeyDataSourceAccessByNameConfig(rInt string) string {
	return fmt.Sprintf(`
resource "scalr_ssh_key" "test" {
  name         = "ssh-key-test-%s"
  private_key  = <<EOF
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEICNioyJgilYaHbT8pgDXn3haYU0dsl6KJTIvrZm+nIU6
-----END PRIVATE KEY-----
EOF
  account_id   = "%s"
  environments = ["env-svrcnchebt61e30"]
}

data "scalr_ssh_key" "test_by_name" {
  name       = scalr_ssh_key.test.name
  account_id = scalr_ssh_key.test.account_id
}`, rInt, defaultAccount)
}

func testAccScalrSSHKeyDataSourceNotFoundByNameConfig() string {
	return `
data scalr_ssh_key test {
  name       = "ssh-key-foo-bar-baz"
  account_id = "acc-incorrect"
}`
}

func testAccScalrSSHKeyDataSourceMismatchIDNameConfig(rInt string) string {
	return fmt.Sprintf(`
resource "scalr_ssh_key" "test" {
  name         = "ssh-key-test-%s"
  private_key  = <<EOF
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEICNioyJgilYaHbT8pgDXn3haYU0dsl6KJTIvrZm+nIU6
-----END PRIVATE KEY-----
EOF
  account_id   = "%s"
  environments = ["env-svrcnchebt61e30"]
}

data "scalr_ssh_key" "test_mismatch" {
  id         = scalr_ssh_key.test.id
  name       = "incorrect-name"
  account_id = scalr_ssh_key.test.account_id
}`, rInt, defaultAccount)
}
