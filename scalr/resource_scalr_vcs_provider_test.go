package scalr

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccVcsProvider_basic(t *testing.T) {
	provider := &scalr.VcsProvider{}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testVcsAccGithubTokenPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVcsProviderConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVcsProviderExists("scalr_vcs_provider.test", provider),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "name", "github-vcs-provider"),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "vcs_type", string(scalr.Github)),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "url", "https://github.com"),
				),
			},
			{
				Config: testAccScalrVcsProviderUpdate(GITHUB_TOKEN, scalr.Github),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVcsProviderExists("scalr_vcs_provider.test", provider),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "name", "updated-github-vcs-provider"),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "account_id", defaultAccount),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "vcs_type", string(scalr.Github)),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "url", "https://github.com"),
				),
			},
			{
				Config:      testAccScalrVcsProviderUpdate("invalid token", scalr.Github),
				ExpectError: regexp.MustCompile("Invalid access token"),
			},
			{
				Config:      testAccScalrVcsProviderUpdate(GITHUB_TOKEN, scalr.Gitlab),
				ExpectError: regexp.MustCompile("Invalid access token"),
			},
		},
	})
}

func TestAccVcsProvider_globalScope(t *testing.T) {
	provider := &scalr.VcsProvider{}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testVcsAccGithubTokenPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVcsProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scalr_vcs_provider" "test" {
						name="global-github-vcs-provider"
						vcs_type="github"
                        token="%s"
					}
				`, GITHUB_TOKEN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrVcsProviderExists("scalr_vcs_provider.test", provider),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "name", "global-github-vcs-provider"),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "vcs_type", string(scalr.Github)),
					resource.TestCheckResourceAttr("scalr_vcs_provider.test", "url", "https://github.com"),
				),
			},
		},
	})
}

func TestAccScalrVcsProvider_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testVcsAccGithubTokenPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrVcsProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrVcsProviderConfig(),
			},
			{
				ResourceName:            "scalr_vcs_provider.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccCheckScalrVcsProviderExists(resId string, vcsProvider *scalr.VcsProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return noInstanceIdErr
		}

		// Get the role
		p, err := scalrClient.VcsProviders.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*vcsProvider = *p

		return nil
	}
}

func testAccCheckScalrVcsProviderDestroy(s *terraform.State) error {
	scalrClient := testAccProvider.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_vcs_provider" {
			continue
		}

		if rs.Primary.ID == "" {
			return noInstanceIdErr
		}

		_, err := scalrClient.VcsProviders.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Role %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrVcsProviderConfig() string {
	return fmt.Sprintf(`
resource "scalr_vcs_provider" "test" {
  name           = "github-vcs-provider"
  account_id     = "%s"
  vcs_type="github"
  token = "%s"
}`, defaultAccount, GITHUB_TOKEN)
}

func testAccScalrVcsProviderUpdate(token string, vcsType scalr.VcsType) string {

	return fmt.Sprintf(`
resource "scalr_vcs_provider" "test" {
  name           = "updated-github-vcs-provider"
  account_id     = "%s"
  vcs_type="%s"
  token = "%s"
}`, defaultAccount, string(vcsType), token)
}
