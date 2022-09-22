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

func testResourceScalrVariableStateDataCategoryV0() map[string]interface{} {
	return map[string]interface{}{
		"category": "env",
	}
}

func testResourceScalrVariableStateDataCategoryV1() map[string]interface{} {
	return map[string]interface{}{
		"category": "shell",
	}
}

func TestResourceScalrVariableStateUpgradeV1(t *testing.T) {
	expected := testResourceScalrVariableStateDataCategoryV1()
	actual, err := resourceScalrVariableStateUpgradeV1(testResourceScalrVariableStateDataCategoryV0(), nil)
	assertCorrectState(t, err, actual, expected)
}

func testResourceScalrVariableStateDataDescriptionV1(varID string) map[string]interface{} {
	return map[string]interface{}{
		"id": varID,
	}
}

func testResourceScalrVariableStateDataDescriptionV2(varID string) map[string]interface{} {
	return map[string]interface{}{
		"id":          varID,
		"description": "",
	}
}

func TestResourceScalrVariableStateUpgradeV2(t *testing.T) {
	client, err := createScalrClient()
	if err != nil {
		t.Fatal(err)
	}
	variable, err := client.Variables.Create(ctx, scalr.VariableCreateOptions{
		Key:         scalr.String("boo"),
		Value:       scalr.String("bar"),
		Description: scalr.String(""),
		Category:    scalr.Category("shell"),
		HCL:         scalr.Bool(false),
		Sensitive:   scalr.Bool(false),
	})
	defer func() {
		if err := client.Variables.Delete(ctx, variable.ID); err != nil {
			t.Errorf("Error destroying variable! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Variable: %s\nError: %s", variable.Key, err)
		}
	}()

	if err != nil {
		t.Fatal(err)
	}
	expected := testResourceScalrVariableStateDataDescriptionV2(variable.ID)
	actual, err := resourceScalrVariableStateUpgradeV2(testResourceScalrVariableStateDataDescriptionV1(variable.ID), client)
	assertCorrectState(t, err, actual, expected)
}
