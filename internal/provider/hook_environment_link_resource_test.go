package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrHookEnvironmentLinkResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("test-hook-env-link")
	resourceName := "scalr_hook_environment_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrHookEnvironmentLinkConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "hook_id"),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "events.0", "pre-plan"),
					resource.TestCheckResourceAttr(resourceName, "events.1", "post-apply"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScalrHookEnvironmentLinkConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "events.0", "pre-init"),
					resource.TestCheckResourceAttr(resourceName, "events.1", "post-plan"),
					resource.TestCheckResourceAttr(resourceName, "events.2", "pre-apply"),
				),
			},
		},
	})
}

func TestAccScalrHookEnvironmentLinkResource_allEvents(t *testing.T) {
	rName := acctest.RandomWithPrefix("test-hook-env-link-all")
	resourceName := "scalr_hook_environment_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrHookEnvironmentLinkConfigAllEvents(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "hook_id"),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					// Check that the "*" value is stored in the state
					resource.TestCheckResourceAttr(resourceName, "events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "events.0", "*"),
				),
			},
		},
	})
}

// Test for the uniqueness validation in events
func TestAccScalrHookEnvironmentLinkResource_uniqueEvents(t *testing.T) {
	rName := acctest.RandomWithPrefix("test-hook-env-link-uniq")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrHookEnvironmentLinkConfigDuplicateEvents(rName),
				ExpectError: regexp.MustCompile(`This attribute contains duplicate values`),
			},
		},
	})
}

func testAccScalrHookEnvironmentLinkConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s"
  account_id = "%[2]s"
}

resource "scalr_vcs_provider" "test" {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token    = "%s"
  account_id = "%[2]s"
}

resource "scalr_hook" "test" {
  name            = "%[1]s"
  interpreter     = "bash"
  scriptfile_path = "script.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  account_id      = "%[2]s"
  
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}

resource "scalr_hook_environment_link" "test" {
  hook_id        = scalr_hook.test.id
  environment_id = scalr_environment.test.id
  events         = ["pre-plan", "post-apply"]
}
`, name, githubToken, defaultAccount)
}

func testAccScalrHookEnvironmentLinkConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s"
  account_id = "%[2]s"
}

resource "scalr_vcs_provider" "test" {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "%s"
  account_id = "%[2]s"
}

resource "scalr_hook" "test" {
  name            = "%[1]s"
  interpreter     = "bash"
  scriptfile_path = "script.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  account_id      = "%[2]s"
  
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}

resource "scalr_hook_environment_link" "test" {
  hook_id        = scalr_hook.test.id
  environment_id = scalr_environment.test.id
  events         = ["pre-init", "post-plan", "pre-apply"]
}
`, name, githubToken, defaultAccount)
}

func testAccScalrHookEnvironmentLinkConfigAllEvents(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s"
  account_id = "%[2]s"
}

resource "scalr_vcs_provider" "test" {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "%s"
  account_id = "%[2]s"
}

resource "scalr_hook" "test" {
  name            = "%[1]s"
  interpreter     = "bash"
  scriptfile_path = "script.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  account_id      = "%[2]s"
  
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}

resource "scalr_hook_environment_link" "test" {
  hook_id        = scalr_hook.test.id
  environment_id = scalr_environment.test.id
  events         = ["*"]
}
`, name, githubToken, defaultAccount)
}

func testAccScalrHookEnvironmentLinkConfigDuplicateEvents(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s"
  account_id = "%[2]s"
}

resource "scalr_vcs_provider" "test" {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "%s"
  account_id = "%[2]s"
}

resource "scalr_hook" "test" {
  name            = "%[1]s"
  interpreter     = "bash"
  scriptfile_path = "script.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  account_id      = "%[2]s"
  
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}

resource "scalr_hook_environment_link" "test" {
  hook_id        = scalr_hook.test.id
  environment_id = scalr_environment.test.id
  events         = ["pre-plan", "pre-plan", "post-apply"]  # Duplicate event should cause validation error
}
`, name, githubToken, defaultAccount)
}
