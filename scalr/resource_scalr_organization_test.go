// +build envtest

package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccTFEOrganization_basic(t *testing.T) {
	org := &scalr.Organization{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"scalr_organization.foobar", org),
					testAccCheckTFEOrganizationAttributes(org),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "collaborator_auth_policy", "password"),
				),
			},
		},
	})
}

func TestAccTFEOrganization_update(t *testing.T) {
	org := &scalr.Organization{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"scalr_organization.foobar", org),
					testAccCheckTFEOrganizationAttributes(org),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "name", "tst-terraform"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "email", "admin@company.com"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "collaborator_auth_policy", "password"),
				),
			},

			{
				Config: testAccTFEOrganization_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationExists(
						"scalr_organization.foobar", org),
					testAccCheckTFEOrganizationAttributesUpdated(org),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "name", "terraform-updated"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "email", "admin-updated@company.com"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "session_timeout_minutes", "3600"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "session_remember_minutes", "3600"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "collaborator_auth_policy", "password"),
					resource.TestCheckResourceAttr(
						"scalr_organization.foobar", "owners_team_saml_role_id", "owners"),
				),
			},
		},
	})
}

func TestAccTFEOrganization_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganization_basic,
			},

			{
				ResourceName:      "scalr_organization.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEOrganizationExists(
	n string, org *scalr.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		o, err := scalrClient.Organizations.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if o.Name != rs.Primary.ID {
			return fmt.Errorf("Organization not found")
		}

		*org = *o

		return nil
	}
}

func testAccCheckTFEOrganizationAttributes(
	org *scalr.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.Name != "tst-terraform" {
			return fmt.Errorf("Bad name: %s", org.Name)
		}

		if org.Email != "admin@company.com" {
			return fmt.Errorf("Bad email: %s", org.Email)
		}

		if org.CollaboratorAuthPolicy != scalr.AuthPolicyPassword {
			return fmt.Errorf("Bad auth policy: %s", org.CollaboratorAuthPolicy)
		}

		return nil
	}
}

func testAccCheckTFEOrganizationAttributesUpdated(
	org *scalr.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if org.Name != "terraform-updated" {
			return fmt.Errorf("Bad name: %s", org.Name)
		}

		if org.Email != "admin-updated@company.com" {
			return fmt.Errorf("Bad email: %s", org.Email)
		}

		if org.SessionTimeout != 3600 {
			return fmt.Errorf("Bad session timeout minutes: %d", org.SessionTimeout)
		}

		if org.SessionRemember != 3600 {
			return fmt.Errorf("Bad session remember minutes: %d", org.SessionRemember)
		}

		if org.CollaboratorAuthPolicy != scalr.AuthPolicyPassword {
			return fmt.Errorf("Bad auth policy: %s", org.CollaboratorAuthPolicy)
		}

		if org.OwnersTeamSAMLRoleID != "owners" {
			return fmt.Errorf("Bad owners team SAML role ID: %s", org.OwnersTeamSAMLRoleID)
		}

		return nil
	}
}

func testAccCheckTFEOrganizationDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_organization" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Organizations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Organization %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEOrganization_basic = `
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}`

const testAccTFEOrganization_update = `
resource "scalr_organization" "foobar" {
  name                     = "terraform-updated"
  email                    = "admin-updated@company.com"
  session_timeout_minutes  = 3600
  session_remember_minutes = 3600
  owners_team_saml_role_id = "owners"
}`
