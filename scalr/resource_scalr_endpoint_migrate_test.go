package scalr

import (
	"reflect"
	"testing"
)

func testResourceScalrEndpointStateDataV0V1() map[string]interface{} {
	return map[string]interface{}{
		"name":           "test-name",
		"timeout":        3,
		"environment_id": "env-123",
	}
}

func TestResourceScalrEndpointStateUpgradeV0(t *testing.T) {
	expected := testResourceScalrEndpointStateDataV0V1()
	//ensure that migration doesn't affect state
	actual, err := resourceScalrEndpointStateUpgradeV0(testResourceScalrEndpointStateDataV0V1(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
