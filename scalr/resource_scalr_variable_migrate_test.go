package scalr

import (
	"reflect"
	"testing"

	scalr "github.com/scalr/go-scalr"
)

func testResourceScalrVariableStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"workspace_id": "my-env/a-workspace",
	}
}

func testResourceScalrVariableStateDataV1() map[string]interface{} {
	return map[string]interface{}{
		"workspace_id": "ws-123",
	}
}

func TestResourceScalrVariableStateUpgradeV0(t *testing.T) {
	client := testScalrClient(t)
	name := "a-workspace"
	client.Workspaces.Create(nil, scalr.WorkspaceCreateOptions{
		ID:          "ws-123",
		Name:        &name,
		Environment: &scalr.Environment{ID: "my-env"},
	})

	expected := testResourceScalrVariableStateDataV1()
	actual, err := resourceScalrVariableStateUpgradeV0(testResourceScalrVariableStateDataV0(), client)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
