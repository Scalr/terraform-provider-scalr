package scalr

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalrAccountAllowedIps_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
		},
	})
}

func TestAccScalrAccountAllowedIps_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
	rg, _ := regexp.Compile(`config is invalid: allowed_ips: attribute supports 1 item as a minimum, config has 0 declared`)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
