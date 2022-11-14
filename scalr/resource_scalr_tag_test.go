package scalr

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scalr/go-scalr"
	"log"
	"testing"
)

func TestAccScalrTag_basic(t *testing.T) {
	tag := &scalr.Tag{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrTagDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrTagDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrTagDestroy,
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

func TestAccScalrTag_renamed(t *testing.T) {
	tag := &scalr.Tag{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalrTagDestroy,
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
				PreConfig: testAccCheckScalrTagRename(tag),
				Config:    testAccScalrTagRenamed(),
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalr_tag.test", "name", "renamed-outside-terraform"),
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

func testAccCheckScalrTagRename(tag *scalr.Tag) func() {
	return func() {
		scalrClient := testAccProvider.Meta().(*scalr.Client)

		t, err := scalrClient.Tags.Read(ctx, tag.ID)

		if err != nil {
			log.Fatalf("Error retrieving tag: %v", err)
		}

		t, err = scalrClient.Tags.Update(
			context.Background(),
			t.ID,
			scalr.TagUpdateOptions{Name: scalr.String("renamed-outside-terraform")},
		)

		if err != nil {
			log.Fatalf("Could not rename the tag outside of terraform: %v", err)
		}

		if t.Name != "renamed-outside-terraform" {
			log.Fatalf("Failed to rename the tag outside of terraform: %v", err)
		}
	}
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
