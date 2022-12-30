package scalr

import (
	"testing"
)

func testResourceScalrVcsProviderStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"id":       "vcs-123",
		"name":     "test",
		"token":    "tolen",
		"url":      "https://github.com",
		"vcs_type": "github",
	}
}

func testResourceScalrVcsProviderStateDataV1() map[string]interface{} {
	res := testResourceScalrVcsProviderStateDataV0()
	res["username"] = ""
	return res
}

func TestResourceScalrVcsProviderStateUpgradeV0(t *testing.T) {
	expected := testResourceScalrVcsProviderStateDataV1()
	actual, err := resourceScalrVcsProviderStateUpgradeV0(ctx, testResourceScalrVcsProviderStateDataV0(), nil)
	assertCorrectState(t, err, actual, expected)
}
