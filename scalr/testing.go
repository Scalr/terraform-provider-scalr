package scalr

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	scalr "github.com/scalr/go-scalr"
)

func testScalrClient(t *testing.T) *scalr.Client {
	config := &scalr.Config{
		Token: "not-a-token",
	}

	client, err := scalr.NewClient(config)
	if err != nil {
		t.Fatalf("error creating Scalr client: %v", err)
	}

	client.Workspaces = newMockWorkspaces()

	return client
}

func getResourceIDfromState(resourceID *string, resourceDeclaration string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceDeclaration]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceDeclaration)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		*resourceID = rs.Primary.ID
		return nil
	}
}
