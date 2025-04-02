package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrHook_basic(t *testing.T) {
	rInt := GetRandomInteger()
	resourceName := "scalr_hook.test"
	var hook scalr.Hook

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrHookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrHookBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrHookExists(resourceName, &hook),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("hook-test-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "interpreter", "bash"),
					resource.TestCheckResourceAttr(resourceName, "scriptfile_path", "pre-plan.sh"),
					resource.TestCheckResourceAttrSet(resourceName, "vcs_provider_id"),
					resource.TestCheckResourceAttr(resourceName, "vcs_repo.0.identifier", "scalr/terraform-provider-scalr"),
					resource.TestCheckResourceAttr(resourceName, "vcs_repo.0.branch", "main"),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
				),
			},
		},
	})
}

func TestAccScalrHook_update(t *testing.T) {
	rInt := GetRandomInteger()
	resourceName := "scalr_hook.test"
	var hook scalr.Hook

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrHookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrHookBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrHookExists(resourceName, &hook),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("hook-test-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "interpreter", "bash"),
					resource.TestCheckResourceAttr(resourceName, "scriptfile_path", "pre-plan.sh"),
				),
			},
			{
				Config: testAccScalrHookUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrHookExists(resourceName, &hook),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("hook-test-%d-updated", rInt)),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated hook description"),
					resource.TestCheckResourceAttr(resourceName, "interpreter", "python3"),
					resource.TestCheckResourceAttr(resourceName, "scriptfile_path", "scripts/pre-apply.py"),
					resource.TestCheckResourceAttr(resourceName, "vcs_repo.0.identifier", "scalr/terraform-provider-scalr"),
					resource.TestCheckResourceAttr(resourceName, "vcs_repo.0.branch", "develop"),
				),
			},
		},
	})
}

func TestAccScalrHook_import(t *testing.T) {
	rInt := GetRandomInteger()
	resourceName := "scalr_hook.test"
	var hook scalr.Hook

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrHookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrHookBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrHookExists(resourceName, &hook),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("hook-test-%d", rInt)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalrHookExists(
	n string, hook *scalr.Hook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no hook ID is set")
		}

		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)
		h, err := scalrClient.Hooks.Read(ctx, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading hook %s: %v", rs.Primary.ID, err)
		}

		*hook = *h
		return nil
	}
}

func testAccCheckScalrHookDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_hook" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Hooks.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Hook %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccScalrHookBasic(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_vcs_provider" "test" {
  name     = "vcs-test-%d"
  vcs_type = "github"
  token    = "%s"
}

resource "scalr_hook" "test" {
  name            = "hook-test-%d"
  interpreter     = "bash"
  scriptfile_path = "pre-plan.sh"
  vcs_provider_id = scalr_vcs_provider.test.id
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "main"
  }
}`, rInt, githubToken, rInt)
}

func testAccScalrHookUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "scalr_vcs_provider" "test" {
  name     = "vcs-test-%d"
  vcs_type = "github"
  token    = "%s"
}

resource "scalr_hook" "test" {
  name            = "hook-test-%d-updated"
  description     = "Updated hook description"
  interpreter     = "python3"
  scriptfile_path = "scripts/pre-apply.py"
  vcs_provider_id = scalr_vcs_provider.test.id
  vcs_repo {
    identifier = "scalr/terraform-provider-scalr"
    branch     = "develop"
  }
}`, rInt, githubToken, rInt)
}
