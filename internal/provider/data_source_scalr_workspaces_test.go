package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScalrWorkspacesDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccScalrWorkspacesDataSourceConfigBasic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_workspaces.foobar", "ids.#", "2"),
					resource.TestCheckResourceAttrSet("data.scalr_workspaces.foobar", "id"),
					resource.TestCheckResourceAttr("data.scalr_workspaces.tagged", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.scalr_workspaces.partial", "ids.#", "3"),
				),
			},
		},
	})
}

func testAccScalrWorkspacesDataSourceConfigBasic(rInt int) string {
	return fmt.Sprintf(`
resource scalr_tag foo {
  name = "foo"
}

resource scalr_tag bar {
  name = "bar"
}

resource scalr_tag baz {
  name = "baz"
}

resource scalr_environment env {
  name = "env-test"
}

resource scalr_workspace foo {
  environment_id = scalr_environment.env.id
  name = "env-foo-%d"
}

resource scalr_workspace bar {
  environment_id = scalr_environment.env.id
  name = "env-bar-%[1]d"
}

resource scalr_workspace dummy {
  environment_id = scalr_environment.env.id
  name = "env-dummy-%[1]d"
}

resource scalr_workspace foobar-tagged {
  environment_id = scalr_environment.env.id
  name    = "foobar-tagged-%[1]d"
  tag_ids = [scalr_tag.foo.id, scalr_tag.bar.id]
}

resource scalr_workspace barbaz-tagged {
  environment_id = scalr_environment.env.id
  name    = "barbaz-tagged-%[1]d"
  tag_ids = [scalr_tag.bar.id, scalr_tag.baz.id]
}

resource scalr_workspace baz-tagged {
  environment_id = scalr_environment.env.id
  name    = "baz-tagged-%[1]d"
  tag_ids = [scalr_tag.baz.id]
}

data scalr_workspaces foobar {
  name = "in:${scalr_workspace.foo.name},${scalr_workspace.bar.name}"
}

data scalr_workspaces tagged {
  tag_ids = [scalr_tag.foo.id, scalr_tag.bar.id]
  depends_on = [
    scalr_workspace.foobar-tagged,
	scalr_workspace.barbaz-tagged,
    scalr_workspace.baz-tagged,
  ]
}

data scalr_workspaces partial {
  name       = "like:bar"
  depends_on = [
    scalr_workspace.foo,
    scalr_workspace.bar,
    scalr_workspace.foobar-tagged,
	scalr_workspace.barbaz-tagged,
  ]
}`, rInt)
}
