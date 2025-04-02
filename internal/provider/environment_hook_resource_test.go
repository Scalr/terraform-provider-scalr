package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrEnvironmentHookResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("test-hook-env-link")
	resourceName := "scalr_environment_hook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrEnvironmentHookConfig(rName),
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
				Config: testAccScalrEnvironmentHookConfigUpdated(rName),
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
	resourceName := "scalr_environment_hook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrEnvironmentHookConfigAllEvents(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "hook_id"),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "events.0", "*"),
				),
			},
		},
	})
}

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
				Config:      testAccScalrEnvironmentHookConfigDuplicateEvents(rName),
				ExpectError: regexp.MustCompile(`This attribute contains duplicate values`),
			},
		},
	})
}

func testAccScalrEnvironmentHookConfig(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s"
  account_id = "%[2]s"
}

resource "scalr_vcs_provider" "test" {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token    = "%[3]s"
  account_id = "%[2]s"
}

resource "scalr_hook" "test" {
  name            = "%[1]s"
  interpreter     = "bash"
  scriptfile_path = "script.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}

resource "scalr_environment_hook" "test" {
  hook_id        = scalr_hook.test.id
  environment_id = scalr_environment.test.id
  events         = ["pre-plan", "post-apply"]
}
`, name, defaultAccount, githubToken)
}

func testAccScalrEnvironmentHookConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s"
  account_id = "%[2]s"
}

resource "scalr_vcs_provider" "test" {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "%[3]s"
  account_id = "%[2]s"
}

resource "scalr_hook" "test" {
  name            = "%[1]s"
  interpreter     = "bash"
  scriptfile_path = "script.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}

resource "scalr_environment_hook" "test" {
  hook_id        = scalr_hook.test.id
  environment_id = scalr_environment.test.id
  events         = ["pre-init", "post-plan", "pre-apply"]
}
`, name, defaultAccount, githubToken)
}

func testAccScalrEnvironmentHookConfigAllEvents(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s"
  account_id = "%[2]s"
}

resource "scalr_vcs_provider" "test" {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "%[3]s"
  account_id = "%[2]s"
}

resource "scalr_hook" "test" {
  name            = "%[1]s"
  interpreter     = "bash"
  scriptfile_path = "script.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}

resource "scalr_environment_hook" "test" {
  hook_id        = scalr_hook.test.id
  environment_id = scalr_environment.test.id
  events         = ["*"]
}
`, name, defaultAccount, githubToken)
}

func testAccScalrEnvironmentHookConfigDuplicateEvents(name string) string {
	return fmt.Sprintf(`
resource "scalr_environment" "test" {
  name       = "%[1]s"
  account_id = "%[2]s"
}

resource "scalr_vcs_provider" "test" {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "%[3]s"
  account_id = "%[2]s"
}

resource "scalr_hook" "test" {
  name            = "%[1]s"
  interpreter     = "bash"
  scriptfile_path = "script.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}

resource "scalr_environment_hook" "test" {
  hook_id        = scalr_hook.test.id
  environment_id = scalr_environment.test.id
  events         = ["pre-plan", "pre-plan", "post-apply"]  # Duplicate event should cause validation error
}
`, name, defaultAccount, githubToken)
}
