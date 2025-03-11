package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrHookDataSource_basic(t *testing.T) {
	hookName := acctest.RandomWithPrefix("test-hook")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			t.Skip("Works with personal token but does not work with github action token.")
			testVcsAccGithubTokenPreCheck(t)
		},
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      `data scalr_hook test {}`,
				ExpectError: regexp.MustCompile(`At least one of these attributes must be configured: \[id,name]`),
			},
			{
				Config:      `data scalr_hook test {id = ""}`,
				ExpectError: regexp.MustCompile("Attribute id must not be empty"),
			},
			{
				Config:      `data scalr_hook test {name = ""}`,
				ExpectError: regexp.MustCompile("Attribute name must not be empty"),
			},
			{
				Config: testAccScalrHookDataSourceByIDConfig(hookName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "name", hookName),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "interpreter", "bash"),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "scriptfile_path", "script.sh"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_provider_id"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "account_id"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_repo.0.identifier"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_repo.0.branch"),
				),
			},
			{
				Config: testAccScalrHookDataSourceByNameConfig(hookName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "name", hookName),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "interpreter", "bash"),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "scriptfile_path", "script.sh"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_provider_id"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "account_id"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_repo.0.identifier"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_repo.0.branch"),
				),
			},
			{
				Config: testAccScalrHookDataSourceByIDAndNameConfig(hookName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "id"),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "name", hookName),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "interpreter", "bash"),
					resource.TestCheckResourceAttr("data.scalr_hook.test", "scriptfile_path", "script.sh"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_provider_id"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "account_id"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_repo.0.identifier"),
					resource.TestCheckResourceAttrSet("data.scalr_hook.test", "vcs_repo.0.branch"),
				),
			},
		},
	})
}

func testAccScalrHookDataSourceByIDConfig(name string) string {
	return fmt.Sprintf(`
resource scalr_vcs_provider test {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "token"
}

resource scalr_hook test {
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

data scalr_hook test {
  id         = scalr_hook.test.id
  account_id = "%[2]s"
}`, name, defaultAccount)
}

func testAccScalrHookDataSourceByNameConfig(name string) string {
	return fmt.Sprintf(`
resource scalr_vcs_provider test {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "%s"
}

resource scalr_hook test {
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

data scalr_hook test {
  name       = scalr_hook.test.name
  account_id = "%[2]s"
}`, name, githubToken, defaultAccount)
}

func testAccScalrHookDataSourceByIDAndNameConfig(name string) string {
	return fmt.Sprintf(`
resource scalr_vcs_provider test {
  name       = "%[1]s-vcs"
  vcs_type   = "github"
  token      = "%s"
}

resource scalr_hook test {
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

data scalr_hook test {
  id         = scalr_hook.test.id
  name       = scalr_hook.test.name
  account_id = "%[2]s"
}`, name, githubToken, defaultAccount)
}
