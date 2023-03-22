package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalrEnvironmentIDsDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrEnvironmentIDsDataSourceConfigBasic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_environment_ids.foobar", "names.#", "2"),
					resource.TestCheckResourceAttr(
						"data.scalr_environment_ids.foobar", "names.0", fmt.Sprintf("env-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_environment_ids.foobar", "names.1", fmt.Sprintf("env-bar-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.scalr_environment_ids.foobar", "ids.%", "2"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_environment_ids.foobar", fmt.Sprintf("ids.env-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_environment_ids.foobar", fmt.Sprintf("ids.env-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.scalr_environment_ids.foobar", "id"),
					resource.TestCheckResourceAttr("data.scalr_environment_ids.tagged", "ids.%", "2"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_environment_ids.tagged", fmt.Sprintf("ids.foobar-tagged-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_environment_ids.tagged", fmt.Sprintf("ids.barbaz-tagged-%d", rInt)),
					resource.TestCheckResourceAttr("data.scalr_environment_ids.partial", "ids.%", "4"),
					resource.TestCheckResourceAttrSet(
						"data.scalr_environment_ids.partial", fmt.Sprintf("ids.env-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_environment_ids.partial", fmt.Sprintf("ids.env-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_environment_ids.partial", fmt.Sprintf("ids.foobar-tagged-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.scalr_environment_ids.partial", fmt.Sprintf("ids.barbaz-tagged-%d", rInt)),
				),
			},
		},
	})
}

func testAccScalrEnvironmentIDsDataSourceConfigBasic(rInt int) string {
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

resource scalr_environment foo {
  name = "env-foo-%d"
}

resource scalr_environment bar {
  name = "env-bar-%[1]d"
}

resource scalr_environment dummy {
  name = "env-dummy-%[1]d"
}

resource scalr_environment foobar-tagged {
  name    = "foobar-tagged-%[1]d"
  tag_ids = [scalr_tag.foo.id, scalr_tag.bar.id]
}

resource scalr_environment barbaz-tagged {
  name    = "barbaz-tagged-%[1]d"
  tag_ids = [scalr_tag.bar.id, scalr_tag.baz.id]
}

resource scalr_environment baz-tagged {
  name    = "baz-tagged-%[1]d"
  tag_ids = [scalr_tag.baz.id]
}

data scalr_environment_ids foobar {
  names = [scalr_environment.foo.name, scalr_environment.bar.name]
}

data scalr_environment_ids tagged {
  tag_ids = [scalr_tag.foo.id, scalr_tag.bar.id]
  depends_on = [
    scalr_environment.foobar-tagged,
	scalr_environment.barbaz-tagged,
    scalr_environment.baz-tagged,
  ]
}

data scalr_environment_ids partial {
  names          = ["foo", "bar"]
  exact_match    = false
  depends_on = [
    scalr_environment.foo,
    scalr_environment.bar,
    scalr_environment.foobar-tagged,
	scalr_environment.barbaz-tagged,
  ]
}`, rInt)
}
