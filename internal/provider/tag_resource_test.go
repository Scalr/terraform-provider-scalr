package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrTag_basic(t *testing.T) {
	tagName := acctest.RandomWithPrefix("test-tag")
	tag := &scalr.Tag{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrTagBasic(tagName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrTagExists("scalr_tag.test", tag),
					resource.TestCheckResourceAttr("scalr_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("scalr_tag.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func TestAccScalrTag_import(t *testing.T) {
	tagName := acctest.RandomWithPrefix("test-tag")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrTagBasic(tagName),
			},

			{
				ResourceName:      "scalr_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalrTag_update(t *testing.T) {
	tagName := acctest.RandomWithPrefix("test-tag")
	tagNameUpdated := acctest.RandomWithPrefix("test-tag")
	tag := &scalr.Tag{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckScalrTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrTagBasic(tagName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrTagExists("scalr_tag.test", tag),
					resource.TestCheckResourceAttr("scalr_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("scalr_tag.test", "account_id", defaultAccount),
				),
			},

			{
				Config: testAccScalrTagUpdate(tagNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrTagExists("scalr_tag.test", tag),
					resource.TestCheckResourceAttr("scalr_tag.test", "name", tagNameUpdated),
					resource.TestCheckResourceAttr("scalr_tag.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func testAccScalrTagBasic(name string) string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "%[1]s"
  account_id = "%[2]s"
}`, name, defaultAccount)
}

func testAccScalrTagUpdate(name string) string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "%[1]s"
  account_id = "%[2]s"
}`, name, defaultAccount)
}

func testAccCheckScalrTagExists(resId string, tag *scalr.Tag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

		rs, ok := s.RootModule().Resources[resId]
		if !ok {
			return fmt.Errorf("Not found: %s", resId)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the tag
		t, err := scalrClient.Tags.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*tag = *t

		return nil
	}
}

func testAccCheckScalrTagDestroy(s *terraform.State) error {
	scalrClient := testAccProviderSDK.Meta().(*scalr.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scalr_tag" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := scalrClient.Tags.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Tag %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func TestAccScalrTag_UpgradeFromSDK(t *testing.T) {
	tagName := acctest.RandomWithPrefix("test-tag")

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"scalr": {
						Source:            "registry.scalr.io/scalr/scalr",
						VersionConstraint: "<=2.2.0",
					},
				},
				Config: testAccScalrTagBasic(tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scalr_tag.test", "id"),
					resource.TestCheckResourceAttr("scalr_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("scalr_tag.test", "account_id", defaultAccount),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(t),
				Config:                   testAccScalrTagBasic(tagName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
