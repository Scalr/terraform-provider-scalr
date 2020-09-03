package scalr

import (
	"reflect"
	"testing"
)

func testResourceScalrWorkspaceStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"id":          "my-org/test",
		"external_id": "ws-123",
	}
}

func testResourceScalrWorkspaceStateDataV1() map[string]interface{} {
	v0 := testResourceScalrWorkspaceStateDataV0()
	return map[string]interface{}{
		"id": v0["external_id"],
	}
}

func TestResourceScalrWorkspaceStateUpgradeV0(t *testing.T) {
	expected := testResourceScalrWorkspaceStateDataV1()
	actual, err := resourceScalrWorkspaceStateUpgradeV0(testResourceScalrWorkspaceStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

func testResourceScalrWorkspaceStateDataV1VcsRepo() map[string]interface{} {
	return map[string]interface{}{
		"id":          "my-org/test",
		"vcs_repo":    map[string]interface{}{
			"oauth_token_id": "test_provider_id",
			"identifier": "test_identifier",
		},
	}
}

func testResourceScalrWorkspaceStateDataV2() map[string]interface{} {
	v1 := testResourceScalrWorkspaceStateDataV1VcsRepo()
	vcsRepo := v1["vcs_repo"].(map[string]interface{})
	return map[string]interface{}{
		"id":          v1["id"],
		"vcs_repo":    map[string]interface{}{
			"identifier": vcsRepo["identifier"],
		},
		"vcs_provider_id": vcsRepo["oauth_token_id"],
	}
}

func TestResourceScalrWorkspaceStateUpgradeV1(t *testing.T) {
	expected := testResourceScalrWorkspaceStateDataV2()
	actual, err := resourceScalrWorkspaceStateUpgradeV1(testResourceScalrWorkspaceStateDataV1VcsRepo(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

