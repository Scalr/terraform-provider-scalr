package scalr

import (
	"context"
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
	client.Workspaces.Create(context.Background(), scalr.WorkspaceCreateOptions{
		ID:          "ws-123",
		Name:        &name,
		Environment: &scalr.Environment{ID: "my-env"},
	})

	expected := testResourceScalrVariableStateDataV1()
	actual, err := resourceScalrVariableStateUpgradeV0(testResourceScalrVariableStateDataV0(), client)
	assertCorrectState(t, err, actual, expected)
}

func testResourceScalrVariableStateDataCategoryV2() map[string]interface{} {
	return map[string]interface{}{
		"category": "env",
	}
}

func testResourceScalrVariableStateDataV2() map[string]interface{} {
	return map[string]interface{}{
		"category": "shell",
	}
}

func TestResourceScalrVariableStateUpgradeV1(t *testing.T) {
	expected := testResourceScalrVariableStateDataV2()
	actual, err := resourceScalrVariableStateUpgradeV1(testResourceScalrVariableStateDataCategoryV2(), nil)
	assertCorrectState(t, err, actual, expected)
}
