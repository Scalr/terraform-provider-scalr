package scalr

import (
	"testing"
)

func testResourceScalrEndpointStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"id":          "endpoint-id",
		"http_method": "POST",
	}
}

func testResourceScalrEndpointStateDataV1() map[string]interface{} {
	v0 := testResourceScalrEndpointStateDataV0()
	delete(v0, "http_method")
	return v0
}

func TestResourceScalrEndpointStateUpgradeV0(t *testing.T) {
	expected := testResourceScalrEndpointStateDataV1()
	actual, err := resourceScalrEndpointStateUpgradeV0(ctx, testResourceScalrEndpointStateDataV0(), nil)
	assertCorrectState(t, err, actual, expected)
}
