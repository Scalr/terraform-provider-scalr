package scalr

import (
	"reflect"
	"testing"

	scalr "github.com/scalr/go-scalr"
)

const defaultAccount = "acc-svrcncgh453bi8g"

func testScalrClient(t *testing.T) *scalr.Client {
	config := &scalr.Config{
		Token: "not-a-token",
	}

	client, err := scalr.NewClient(config)
	if err != nil {
		t.Fatalf("error creating Scalr client: %v", err)
	}

	client.Workspaces = newMockWorkspaces()
	client.Variables = newMockVariables()

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
