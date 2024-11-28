package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scalr/go-scalr"
)

func TestAccScalrTag_basic(t *testing.T) {
	tag := &scalr.Tag{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrTagBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrTagExists("scalr_tag.test", tag),
					resource.TestCheckResourceAttr("scalr_tag.test", "name", "test-tag-name"),
					resource.TestCheckResourceAttr("scalr_tag.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func TestAccScalrTag_import(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrTagBasic(),
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
	tag := &scalr.Tag{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckScalrTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalrTagBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrTagExists("scalr_tag.test", tag),
					resource.TestCheckResourceAttr("scalr_tag.test", "name", "test-tag-name"),
					resource.TestCheckResourceAttr("scalr_tag.test", "account_id", defaultAccount),
				),
			},

			{
				Config: testAccScalrTagUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalrTagExists("scalr_tag.test", tag),
					resource.TestCheckResourceAttr("scalr_tag.test", "name", "test-tag-name-updated"),
					resource.TestCheckResourceAttr("scalr_tag.test", "account_id", defaultAccount),
				),
			},
		},
	})
}

func testAccScalrTagBasic() string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "test-tag-name"
  account_id = "%s"
}`, defaultAccount)
}

func testAccScalrTagUpdate() string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "test-tag-name-updated"
  account_id = "%s"
}`, defaultAccount)
}

func testAccScalrTagRenamed() string {
	return fmt.Sprintf(`
resource scalr_tag test {
  name       = "renamed-outside-terraform"
  account_id = "%s"
}`, defaultAccount)
}

func testAccCheckScalrTagExists(resId string, tag *scalr.Tag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

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
	scalrClient := testAccProvider.Meta().(*scalr.Client)

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
