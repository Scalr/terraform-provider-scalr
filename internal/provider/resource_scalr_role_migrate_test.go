package provider

import (
	"testing"
)

func testResourceScalrRoleStateDataV0() map[string]interface{} {
	perms := []interface{}{"accounts:update", "global-scope:read"}
	return map[string]interface{}{
		"permissions": perms,
	}
}

func testResourceScalrRoleStateDataV1() map[string]interface{} {
	perms := []interface{}{"global-scope:read", "accounts:update", "accounts:set-quotas"}
	return map[string]interface{}{
		"permissions": perms,
	}
}

func testResourceScalrRoleStateDataV0NoGlobalScope() map[string]interface{} {
	perms := []interface{}{"accounts:create"}
	return map[string]interface{}{
		"permissions": perms,
	}
}

func testResourceScalrRoleStateDataV1NoGlobalScope() map[string]interface{} {
	perms := []interface{}{"accounts:create"}
	return map[string]interface{}{
		"permissions": perms,
	}
}

func testResourceScalrRoleStateDataV0ExistingPerm() map[string]interface{} {
	perms := []interface{}{"accounts:set-quotas", "global-scope:read"}
	return map[string]interface{}{
		"permissions": perms,
	}
}

func testResourceScalrRoleStateDataV1ExistingPerm() map[string]interface{} {
	perms := []interface{}{"accounts:set-quotas", "global-scope:read"}
	return map[string]interface{}{
		"permissions": perms,
	}
}

func TestResourceScalrRoleStateUpgradeV0(t *testing.T) {
	expected := testResourceScalrRoleStateDataV1()
	actual, err := resourceScalrRoleStateUpgradeV0(ctx, testResourceScalrRoleStateDataV0(), nil)
	assertCorrectState(t, err, actual, expected)
}

func TestResourceScalrRoleStateUpgradeV0NoGlobalScope(t *testing.T) {
	expected := testResourceScalrRoleStateDataV1NoGlobalScope()
	actual, err := resourceScalrRoleStateUpgradeV0(ctx, testResourceScalrRoleStateDataV0NoGlobalScope(), nil)
	assertCorrectState(t, err, actual, expected)
}

func TestResourceScalrRoleStateUpgradeV0ExistingPermission(t *testing.T) {
	expected := testResourceScalrRoleStateDataV1ExistingPerm()
	actual, err := resourceScalrRoleStateUpgradeV0(ctx, testResourceScalrRoleStateDataV0ExistingPerm(), nil)
	assertCorrectState(t, err, actual, expected)
}
