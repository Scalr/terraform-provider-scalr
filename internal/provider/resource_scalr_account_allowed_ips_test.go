package provider

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrAccountAllowedIps_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAccountAllowedIps([]string{"192.168.0.12", "0.0.0.0/0", "192.168.0.0/32"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_account_allowed_ips.test", "id"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.#", "3"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.0", "192.168.0.12"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.1", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.2", "192.168.0.0/32"),
				),
			},
		},
	})
}

func TestAccScalrAccountAllowedIps_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAccountAllowedIps([]string{"192.168.0.12", "0.0.0.0/0"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_account_allowed_ips.test", "id"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.#", "2"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.0", "192.168.0.12"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.1", "0.0.0.0/0"),
				),
			},
			{
				Config: testAccScalrAccountAllowedIps([]string{"0.0.0.0/0"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_account_allowed_ips.test", "id"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.#", "1"),
					resource.TestCheckResourceAttr("scalr_account_allowed_ips.test", "allowed_ips.0", "0.0.0.0/0"),
				),
			},
		},
	})
}

func TestAccScalrAccountAllowedIps_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrAccountAllowedIps([]string{"192.168.0.12", "0.0.0.0/0"}),
			},

			{
				ResourceName:      "scalr_account_allowed_ips.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalrAccountAllowedIps_empty(t *testing.T) {
	rg, _ := regexp.Compile(`Attribute allowed_ips requires 1 item minimum, but config has only 0`)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrAccountAllowedIps([]string{}),
				ExpectError: rg,
			},
		},
	})
}

func TestAccScalrAccountAllowedIps_invalid_CIDR(t *testing.T) {
	rg, _ := regexp.Compile(`value is not a valid IPv4 network`)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrAccountAllowedIps([]string{"192.168.0.12/24"}),
				ExpectError: rg,
			},
		},
	})
}

func testAccScalrAccountAllowedIps(allowedIps []string) string {
	ips, _ := json.Marshal(allowedIps)
	return fmt.Sprintf(`
	resource "scalr_account_allowed_ips" "test" {
		account_id = "%s"
		allowed_ips = %s
	}`, defaultAccount, ips)
}
