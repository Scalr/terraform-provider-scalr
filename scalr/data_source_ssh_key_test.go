package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTFESSHKeyDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESSHKeyDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_ssh_key.foobar", "name", fmt.Sprintf("ssh-key-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_ssh_key.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_ssh_key.foobar", "id"),
				),
			},
		},
	})
}

func testAccTFESSHKeyDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "scalr_ssh_key" "foobar" {
  name         = "ssh-key-test-%d"
  organization = "${scalr_organization.foobar.id}"
  key          = "SSH-KEY-CONTENT"
}

data "scalr_ssh_key" "foobar" {
  name         = "${scalr_ssh_key.foobar.name}"
  organization = "${scalr_ssh_key.foobar.organization}"
}`, rInt, rInt)
}
