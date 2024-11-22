package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrSSHKey_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	var sshKey scalr.SSHKey

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrSSHKeyConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_ssh_key.test", "name", rName),
					testAccCheckSSHKeyIsShared("scalr_ssh_key.test", true, &sshKey),
				),
			},
		},
	})
}

func TestAccScalrSSHKey_update(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rNewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	var sshKey scalr.SSHKey

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrSSHKeyConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_ssh_key.test", "name", rName),
					testAccCheckSSHKeyIsShared("scalr_ssh_key.test", true, &sshKey),
				),
			},
			{
				Config: testAccScalrSSHKeyConfigUpdate(rNewName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_ssh_key.test", "name", rNewName),
					testAccCheckSSHKeyIsShared("scalr_ssh_key.test", false, &sshKey),
				),
			},
		},
	})
}

func TestAccScalrSSHKey_import(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrSSHKeyConfig(rName),
			},
			{
				ResourceName:            "scalr_ssh_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
		},
	})
}

func testAccScalrSSHKeyConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_ssh_key" "test" {
  name         = "%s"
  private_key  = <<-EOF
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBvMDyNaYtWK2TmJIfFhmPZeGxK0bWnNDhjlTZ+V6e4x
-----END PRIVATE KEY-----
EOF
  account_id   = "%s"
  environments = ["*"]
}
`, name, defaultAccount)
}

func testAccScalrSSHKeyConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "scalr_ssh_key" "test" {
  name         = "%s"
  private_key  = <<-EOF
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBvMDyNaYtWK2TmJIfFhmPZeGxK0bWnNDhjlTZ+V6e4x
-----END PRIVATE KEY-----
EOF
  account_id   = "%s"
  environments = []
}
`, name, defaultAccount)
}

func testAccCheckSSHKeyDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_ssh_key" {
			continue
		}

		_, err := scalrClient.SSHKeys.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("SSH key (%s) still exists", rs.Primary.ID)
		}

		if !errors.Is(err, scalr.ErrResourceNotFound) {
			return err
		}
	}

	return nil
}

func testAccCheckSSHKeyIsShared(resourceName string, expectedIsShared bool, sshKey *scalr.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		scalrClient := testAccProvider.Meta().(*scalr.Client)
		readKey, err := scalrClient.SSHKeys.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error reading SSH key: %s", err)
		}

		*sshKey = *readKey

		if readKey.IsShared != expectedIsShared {
			return fmt.Errorf("Expected IsShared to be %t, but got %t", expectedIsShared, readKey.IsShared)
		}

		return nil
	}
}
