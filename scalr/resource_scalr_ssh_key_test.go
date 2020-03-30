package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	tfe "github.com/scalr/go-tfe"
)

func TestAccTFESSHKey_basic(t *testing.T) {
	sshKey := &tfe.SSHKey{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESSHKey_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists(
						"scalr_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributes(sshKey),
					resource.TestCheckResourceAttr(
						"scalr_ssh_key.foobar", "name", "ssh-key-test"),
					resource.TestCheckResourceAttr(
						"scalr_ssh_key.foobar", "key", "SSH-KEY-CONTENT"),
				),
			},
		},
	})
}

func TestAccTFESSHKey_update(t *testing.T) {
	sshKey := &tfe.SSHKey{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFESSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFESSHKey_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists(
						"scalr_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributes(sshKey),
					resource.TestCheckResourceAttr(
						"scalr_ssh_key.foobar", "name", "ssh-key-test"),
					resource.TestCheckResourceAttr(
						"scalr_ssh_key.foobar", "key", "SSH-KEY-CONTENT"),
				),
			},

			{
				Config: testAccTFESSHKey_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFESSHKeyExists(
						"scalr_ssh_key.foobar", sshKey),
					testAccCheckTFESSHKeyAttributesUpdated(sshKey),
					resource.TestCheckResourceAttr(
						"scalr_ssh_key.foobar", "name", "ssh-key-updated"),
					resource.TestCheckResourceAttr(
						"scalr_ssh_key.foobar", "key", "UPDATED-SSH-KEY-CONTENT"),
				),
			},
		},
	})
}

func testAccCheckTFESSHKeyExists(
	n string, sshKey *tfe.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		sk, err := tfeClient.SSHKeys.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if sk == nil {
			return fmt.Errorf("SSH key not found")
		}

		*sshKey = *sk

		return nil
	}
}

func testAccCheckTFESSHKeyAttributes(
	sshKey *tfe.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sshKey.Name != "ssh-key-test" {
			return fmt.Errorf("Bad name: %s", sshKey.Name)
		}
		return nil
	}
}

func testAccCheckTFESSHKeyAttributesUpdated(
	sshKey *tfe.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sshKey.Name != "ssh-key-updated" {
			return fmt.Errorf("Bad name: %s", sshKey.Name)
		}
		return nil
	}
}

func testAccCheckTFESSHKeyDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_ssh_key" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.SSHKeys.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("SSH key %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFESSHKey_basic = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = "${scalr_organization.foobar.id}"
  key          = "SSH-KEY-CONTENT"
}`

const testAccTFESSHKey_update = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_ssh_key" "foobar" {
  name         = "ssh-key-updated"
  organization = "${scalr_organization.foobar.id}"
  key          = "UPDATED-SSH-KEY-CONTENT"
}`
