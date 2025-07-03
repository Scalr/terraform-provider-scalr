package provider

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/client"
)

const defaultAccount = "acc-svrcncgh453bi8g"
const testUser = "user-svrcmmpcrkmit1g" // tf-admin@scalr.com
const testUserEmail = "tf-admin@scalr.com"
const readOnlyRole = "role-t67mjtmabulckto" // Reader
const userRole = "role-t67mjtmauajto7g"     // User

func testScalrClient(t *testing.T) *scalr.Client {
	config := &scalr.Config{
		Token: "not-a-token",
	}

	scalrClient, err := scalr.NewClient(config)
	if err != nil {
		t.Fatalf("error creating Scalr client: %v", err)
	}

	scalrClient.Workspaces = client.NewMockWorkspaces()
	scalrClient.Variables = client.NewMockVariables()

	return scalrClient
}

func assertCorrectState(t *testing.T, err error, actual, expected map[string]interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

// isAccTest returns true if executed while running acceptance tests
func isAccTest() bool {
	return os.Getenv("TF_ACC") == "1"
}

func createScalrClient() (*scalr.Client, error) {
	config := scalr.DefaultConfig()
	config.Address = fmt.Sprintf("https://%s", os.Getenv(client.HostnameEnvVar))
	scalrClient, err := scalr.NewClient(config)
	return scalrClient, err
}
