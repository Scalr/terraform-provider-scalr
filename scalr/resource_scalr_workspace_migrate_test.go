package scalr

import (
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
	assertCorrectState(t, err, actual, expected)
}

func testResourceScalrWorkspaceStateDataV1VcsRepo() map[string]interface{} {
	return map[string]interface{}{
		"id": "my-org/test",
		"vcs_repo": []interface{}{
			map[string]interface{}{
				"oauth_token_id": "test_provider_id",
				"identifier":     "test_identifier",
			},
		},
	}
}

func testResourceScalrWorkspaceStateDataV2() map[string]interface{} {
	v1 := testResourceScalrWorkspaceStateDataV1VcsRepo()
	vcsRepo := v1["vcs_repo"].([]interface{})[0].(map[string]interface{})
	return map[string]interface{}{
		"id": v1["id"],
		"vcs_repo": []interface{}{
			map[string]interface{}{
				"identifier": "test_identifier",
			},
		},
		"vcs_provider_id": vcsRepo["oauth_token_id"],
	}
}

func testResourceScalrWorkspaceStateDataV2NoVcs() map[string]interface{} {
	v1 := testResourceScalrWorkspaceStateDataV1()
	return map[string]interface{}{
		"id": v1["id"],
	}
}

func TestResourceScalrWorkspaceStateUpgradeV1(t *testing.T) {
	expected := testResourceScalrWorkspaceStateDataV2()
	actual, err := resourceScalrWorkspaceStateUpgradeV1(testResourceScalrWorkspaceStateDataV1VcsRepo(), nil)
	assertCorrectState(t, err, actual, expected)
}

func TestResourceScalrWorkspaceStateUpgradeV1NoVcs(t *testing.T) {
	expected := testResourceScalrWorkspaceStateDataV2NoVcs()
	actual, err := resourceScalrWorkspaceStateUpgradeV1(testResourceScalrWorkspaceStateDataV1(), nil)
	assertCorrectState(t, err, actual, expected)
}

func testResourceScalrWorkspaceStateDataV3() map[string]interface{} {
	v2 := testResourceScalrWorkspaceStateDataV2()
	delete(v2, "queue_all_runs")
	return v2
}

func TestResourceScalrWorkspaceStateUpgradeV2(t *testing.T) {
	expected := testResourceScalrWorkspaceStateDataV3()
	actual, err := resourceScalrWorkspaceStateUpgradeV2(testResourceScalrWorkspaceStateDataV2(), nil)
	assertCorrectState(t, err, actual, expected)
}

func TestResourceScalrWorkspaceStateUpgradeV3(t *testing.T) {
	parameters := []struct {
		actual   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			actual: map[string]interface{}{
				"id":                "ws-123",
				"working_directory": "",
				"vcs_repo": []interface{}{
					map[string]interface{}{
						"identifier":       "test/scalr",
						"branch":           "main",
						"path":             "workspace",
						"trigger_prefixes": []interface{}{},
					},
				},
			},
			expected: map[string]interface{}{
				"id":                "ws-123",
				"working_directory": "workspace",
				"vcs_repo": []interface{}{
					map[string]interface{}{
						"identifier":       "test/scalr",
						"branch":           "main",
						"trigger_prefixes": []interface{}{"workspace"},
					},
				},
			},
		},
		{
			actual: map[string]interface{}{
				"id":                "ws-123",
				"working_directory": "foo",
				"vcs_repo": []interface{}{
					map[string]interface{}{
						"identifier":       "test/scalr",
						"branch":           "main",
						"path":             "workspace",
						"trigger_prefixes": []interface{}{},
					},
				},
			},
			expected: map[string]interface{}{
				"id":                "ws-123",
				"working_directory": "workspace/foo",
				"vcs_repo": []interface{}{
					map[string]interface{}{
						"identifier":       "test/scalr",
						"branch":           "main",
						"trigger_prefixes": []interface{}{"workspace", "workspace/foo"},
					},
				},
			},
		},
		{
			actual: map[string]interface{}{
				"id":                "ws-123",
				"working_directory": "",
				"vcs_repo": []interface{}{
					map[string]interface{}{
						"identifier":       "test/scalr",
						"branch":           "main",
						"path":             "workspace",
						"trigger_prefixes": []interface{}{"foo", "bar"},
					},
				},
			},
			expected: map[string]interface{}{
				"id":                "ws-123",
				"working_directory": "workspace",
				"vcs_repo": []interface{}{
					map[string]interface{}{
						"identifier":       "test/scalr",
						"branch":           "main",
						"trigger_prefixes": []interface{}{"workspace/foo", "workspace/bar"},
					},
				},
			},
		},
	}
	for _, data := range parameters {
		actual, err := resourceScalrWorkspaceStateUpgradeV3(data.actual, nil)
		assertCorrectState(t, err, actual, data.expected)
	}
}
