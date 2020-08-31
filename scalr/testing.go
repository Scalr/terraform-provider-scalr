package scalr

import (
	"testing"

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
