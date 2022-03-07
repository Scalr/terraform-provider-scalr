package scalr

import (
	"testing"
)

func testResourceScalrRoleStateDataV0() map[string]interface{} {
	perms := make([]interface{}, 0)
	return map[string]interface{}{
		"permissions": append(perms, "accounts:update", "global-scope:read"),
	}
}

func testResourceScalrRoleStateDataV1() map[string]interface{} {
	perms := make([]interface{}, 0)
	return map[string]interface{}{
		"permissions": append(perms, "accounts:update", "global-scope:read", "accounts:set-quotas"),
	}
}

func testResourceScalrRoleStateDataV0NoGlobalScope() map[string]interface{} {
	perms := make([]interface{}, 0)
	return map[string]interface{}{
		"permissions": append(perms, "accounts:create"),
	}
}

func testResourceScalrRoleStateDataV1NoGlobalScope() map[string]interface{} {
	perms := make([]interface{}, 0)
	return map[string]interface{}{
		"permissions": append(perms, "accounts:create"),
	}
}

func testResourceScalrRoleStateDataV0ExistingPerm() map[string]interface{} {
	perms := make([]interface{}, 0)
	return map[string]interface{}{
		"permissions": append(perms, "accounts:set-quotas", "global-scope:read"),
	}
}

func testResourceScalrRoleStateDataV1ExistingPerm() map[string]interface{} {
	perms := make([]interface{}, 0)
	return map[string]interface{}{
		"permissions": append(perms, "accounts:set-quotas", "global-scope:read"),
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
