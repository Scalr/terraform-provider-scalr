package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrEnvironmentHookResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("test-hook-env-link")
	resourceName := "scalr_environment_hook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
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
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "pre-plan"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "post-apply"),
				),
			},
			{
				RefreshState: true,
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
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "pre-init"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "post-plan"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "pre-apply"),
				),
			},
		},
	})
}

func TestAccScalrEnvironmentHookResource_allEvents(t *testing.T) {
	rName := acctest.RandomWithPrefix("test-hook-env-link-all")
	resourceName := "scalr_environment_hook.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
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
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "*"),
				),
			},
			{
				RefreshState: true,
			},
			{
				Config: testAccScalrEnvironmentHookConfigAllEventsList(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "pre-init"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "pre-plan"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "post-plan"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "pre-apply"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "post-apply"),
				),
			},
			{
				RefreshState: true,
			},
			{
				Config: testAccScalrEnvironmentHookConfigAllEvents(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "hook_id"),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "events.*", "*"),
				),
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

func testAccScalrEnvironmentHookConfigAllEventsList(name string) string {
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
  events         = ["pre-init", "pre-plan", "post-plan", "pre-apply", "post-apply"]
}
`, name, defaultAccount, githubToken)
}
