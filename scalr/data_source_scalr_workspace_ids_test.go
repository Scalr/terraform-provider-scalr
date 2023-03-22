package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalrWorkspaceIDsDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceIDsDataSourceConfigBasic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.#", "2"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.0", fmt.Sprintf("workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.1", fmt.Sprintf("workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", "environment_id"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "ids.%", "2"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_workspace_ids.foobar", "id"),
				),
			},
		},
	})
}

func TestAccScalrWorkspaceIDsDataSource_wildcard(t *testing.T) {
	t.Skip("Wildcard test is not passing for unknown reasons. Using the wildcard symbol produces no workspaces")
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspaceIDsDataSourceConfigWildcard(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "names.0", "*"),
					resource.TestCheckResourceAttr(
						"data.scalr_workspace_ids.foobar", "ids.%", "3"),
					resource.TestCheckResourceAttrSet("data.scalr_workspace_ids.foobar", "environment_id"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.foobar", fmt.Sprintf("ids.workspace-dummy-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_workspace_ids.foobar", "id"),
					resource.TestCheckResourceAttr("data.scalr_workspace_ids.tagged", "ids.%", "2"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.tagged", fmt.Sprintf("ids.foobar-tagged-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.tagged", fmt.Sprintf("ids.barbaz-tagged-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_workspace_ids.partial", "ids.%", "4"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.partial", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.partial", fmt.Sprintf("ids.workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.partial", fmt.Sprintf("ids.foobar-tagged-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_workspace_ids.partial", fmt.Sprintf("ids.barbaz-tagged-%d", rInt)),
				),
			},
		},
	})
}

func testAccScalrWorkspaceIDsDataSourceConfigBasic(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_tag foo {
  name = "foo"
}

resource scalr_tag bar {
  name = "bar"
}

resource scalr_tag baz {
  name = "baz"
}

resource scalr_workspace foo {
  name           = "workspace-foo-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace bar {
  name           = "workspace-bar-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace dummy {
  name           = "workspace-dummy-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace foobar-tagged {
  name           = "foobar-tagged-%[1]d"
  environment_id = scalr_environment.test.id
  tag_ids        = [scalr_tag.foo.id, scalr_tag.bar.id]
}

resource scalr_workspace barbaz-tagged {
  name           = "barbaz-tagged-%[1]d"
  environment_id = scalr_environment.test.id
  tag_ids        = [scalr_tag.bar.id, scalr_tag.baz.id]
}

resource scalr_workspace baz-tagged {
  name           = "baz-tagged-%[1]d"
  environment_id = scalr_environment.test.id
  tag_ids        = [scalr_tag.baz.id]
}

data scalr_workspace_ids foobar {
  names          = [scalr_workspace.foo.name, scalr_workspace.bar.name]
  environment_id = scalr_environment.test.id
}

data scalr_workspace_ids tagged {
  tag_ids        = [scalr_tag.foo.id, scalr_tag.bar.id]
  environment_id = scalr_environment.test.id
}

data scalr_workspace_ids partial {
  names          = ["foo", "bar"]
  environment_id = scalr_environment.test.id
  exact_match    = false
}`, rInt, defaultAccount)
}

// nolint:unused
func testAccScalrWorkspaceIDsDataSourceConfigWildcard(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace foo {
  name           = "workspace-foo-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace bar {
  name           = "workspace-bar-%[1]d"
  environment_id = scalr_environment.test.id
}

resource scalr_workspace dummy {
  name           = "workspace-dummy-%[1]d"
  environment_id = scalr_environment.test.id
}

data scalr_workspace_ids foobar {
  names          = ["*"]
  environment_id = scalr_environment.test.id
}`, rInt, defaultAccount)
}
