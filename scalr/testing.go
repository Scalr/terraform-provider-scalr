package scalr

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	scalr "github.com/scalr/go-scalr"
)

const defaultAccount = "acc-svrcncgh453bi8g"
const testUser = "user-suh84u6vuvidtbg" // test@scalr.com
const testUserEmail = "test@scalr.com"
const readOnlyRole = "role-t67mjtmabulckto" // Reader
const userRole = "role-t67mjtmauajto7g"     // User

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
	config.Address = fmt.Sprintf("https://%s", os.Getenv("SCALR_HOSTNAME"))
	scalrClient, err := scalr.NewClient(config)
	return scalrClient, err
}
