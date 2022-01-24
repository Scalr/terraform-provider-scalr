package scalr

import (
	"testing"
)

func testResourceScalrRoleStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"permissions": []string{"accounts:create", "global-scope:*"},
	}
}

func testResourceScalrRoleStateDataV1() map[string]interface{} {
	return map[string]interface{}{
		"permissions": []string{"accounts:create", "accounts:set-quotas", "global-scope:*"},
	}
}

func testResourceScalrRoleStateDataV0NoGlobalScope() map[string]interface{} {
	return map[string]interface{}{
		"permissions": []string{"accounts:create"},
	}
}

func testResourceScalrRoleStateDataV1NoGlobalScope() map[string]interface{} {
	return map[string]interface{}{
		"permissions": []string{"accounts:create"},
	}
}

func testResourceScalrRoleStateDataV0ExistingPerm() map[string]interface{} {
	return map[string]interface{}{
		"permissions": []string{"accounts:set-quotas", "global-scope:*"},
	}
}

func testResourceScalrRoleStateDataV1ExistingPerm() map[string]interface{} {
	return map[string]interface{}{
		"permissions": []string{"accounts:set-quotas", "global-scope:*"},
	}
}

func TestResourceScalrRoleStateUpgradeV0(t *testing.T) {
	expected := testResourceScalrRoleStateDataV1()
	actual, err := resourceScalrRoleStateUpgradeV0(testResourceScalrRoleStateDataV0(), nil)
	assertCorrectState(t, err, actual, expected)
}

func TestResourceScalrRoleStateUpgradeV0NoGlobalScope(t *testing.T) {
	expected := testResourceScalrRoleStateDataV1NoGlobalScope()
	actual, err := resourceScalrRoleStateUpgradeV0(testResourceScalrRoleStateDataV0NoGlobalScope(), nil)
	assertCorrectState(t, err, actual, expected)
}

func TestResourceScalrRoleStateUpgradeV0ExistingPermission(t *testing.T) {
	expected := testResourceScalrRoleStateDataV1ExistingPerm()
	actual, err := resourceScalrRoleStateUpgradeV0(testResourceScalrRoleStateDataV0ExistingPerm(), nil)
	assertCorrectState(t, err, actual, expected)
}
