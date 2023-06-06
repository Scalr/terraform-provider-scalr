package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalrEnvironmentsDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrEnvironmentsDataSourceConfigBasic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scalr_environments.foobar", "ids.#", "2"),
					resource.TestCheckResourceAttrSet("data.scalr_environments.foobar", "id"),
					resource.TestCheckResourceAttr("data.scalr_environments.tagged", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.scalr_environments.partial", "ids.#", "3"),
				),
			},
		},
	})
}

func testAccScalrEnvironmentsDataSourceConfigBasic(rInt int) string {
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

data scalr_environments foobar {
  name = "in:${scalr_environment.foo.name},${scalr_environment.bar.name}"
}

data scalr_environments tagged {
  tag_ids = [scalr_tag.foo.id, scalr_tag.bar.id]
  depends_on = [
    scalr_environment.foobar-tagged,
	scalr_environment.barbaz-tagged,
    scalr_environment.baz-tagged,
  ]
}

data scalr_environments partial {
  name       = "like:bar"
  depends_on = [
    scalr_environment.foo,
    scalr_environment.bar,
    scalr_environment.foobar-tagged,
	scalr_environment.barbaz-tagged,
  ]
}`, rInt)
}
