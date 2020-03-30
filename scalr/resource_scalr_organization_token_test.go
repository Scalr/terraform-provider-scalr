package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	tfe "github.com/scalr/go-tfe"
)

func TestAccTFEOrganizationToken_basic(t *testing.T) {
	token := &tfe.OrganizationToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"scalr_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"scalr_organization_token.foobar", "organization", "tst-terraform"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationToken_existsWithoutForce(t *testing.T) {
	token := &tfe.OrganizationToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"scalr_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"scalr_organization_token.foobar", "organization", "tst-terraform"),
				),
			},

			{
				Config:      testAccTFEOrganizationToken_existsWithoutForce,
				ExpectError: regexp.MustCompile(`token already exists`),
			},
		},
	})
}

func TestAccTFEOrganizationToken_existsWithForce(t *testing.T) {
	token := &tfe.OrganizationToken{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"scalr_organization_token.foobar", token),
					resource.TestCheckResourceAttr(
						"scalr_organization_token.foobar", "organization", "tst-terraform"),
				),
			},

			{
				Config: testAccTFEOrganizationToken_existsWithForce,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationTokenExists(
						"scalr_organization_token.regenerated", token),
					resource.TestCheckResourceAttr(
						"scalr_organization_token.regenerated", "organization", "tst-terraform"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationToken_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationToken_basic,
			},

			{
				ResourceName:            "scalr_organization_token.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccCheckTFEOrganizationTokenExists(
	n string, token *tfe.OrganizationToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ot, err := tfeClient.OrganizationTokens.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if ot == nil {
			return fmt.Errorf("OrganizationToken not found")
		}

		*token = *ot

		return nil
	}
}

func testAccCheckTFEOrganizationTokenDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_organization_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.OrganizationTokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("OrganizationToken %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEOrganizationToken_basic = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_organization_token" "foobar" {
  organization = "${scalr_organization.foobar.id}"
}`

const testAccTFEOrganizationToken_existsWithoutForce = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_organization_token" "foobar" {
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_organization_token" "error" {
  organization = "${scalr_organization.foobar.id}"
}`

const testAccTFEOrganizationToken_existsWithForce = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_organization_token" "foobar" {
  organization = "${scalr_organization.foobar.id}"
}

resource "scalr_organization_token" "regenerated" {
  organization     = "${scalr_organization.foobar.id}"
  force_regenerate = true
}`
