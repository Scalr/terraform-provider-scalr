package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func TestAccTFEOAuthClient_basic(t *testing.T) {
	oc := &scalr.OAuthClient{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if GITHUB_TOKEN == "" {
				t.Skip("Please set GITHUB_TOKEN to run this test")
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOAuthClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOAuthClient_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOAuthClientExists("scalr_oauth_client.foobar", oc),
					testAccCheckTFEOAuthClientAttributes(oc),
					resource.TestCheckResourceAttr(
						"scalr_oauth_client.foobar", "api_url", "https://api.github.com"),
					resource.TestCheckResourceAttr(
						"scalr_oauth_client.foobar", "http_url", "https://github.com"),
					resource.TestCheckResourceAttr(
						"scalr_oauth_client.foobar", "service_provider", "github"),
				),
			},
		},
	})
}

func testAccCheckTFEOAuthClientExists(
	n string, oc *scalr.OAuthClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		client, err := scalrClient.OAuthClients.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if client.ID != rs.Primary.ID {
			return fmt.Errorf("OAuth client not found")
		}

		*oc = *client

		return nil
	}
}

func testAccCheckTFEOAuthClientAttributes(
	oc *scalr.OAuthClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if oc.APIURL != "https://api.github.com" {
			return fmt.Errorf("Bad API URL: %s", oc.APIURL)
		}

		if oc.HTTPURL != "https://github.com" {
			return fmt.Errorf("Bad HTTP URL: %s", oc.HTTPURL)
		}

		if oc.ServiceProvider != scalr.ServiceProviderGithub {
			return fmt.Errorf("Bad service provider: %s", oc.ServiceProvider)
		}

		return nil
	}
}

func testAccCheckTFEOAuthClientDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_oauth_client" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.OAuthClients.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("OAuth client %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

var testAccTFEOAuthClient_basic = fmt.Sprintf(`
resource "scalr_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "scalr_oauth_client" "foobar" {
  organization     = "${scalr_organization.foobar.id}"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}`, GITHUB_TOKEN)
