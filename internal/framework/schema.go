package framework

import (
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/scalr/go-scalr/v2/scalr/value"
)

func HashString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return 0
}

// SetIfKnownString returns a value.Value that is set to the given string if it is not null or unknown.
// Returns an unset value otherwise.
func SetIfKnownString(v basetypes.StringValue) *value.Value[string] {
	if v.IsNull() || v.IsUnknown() {
		return value.Unset[string]()
	}
	return value.Set(v.ValueString())
}

// SetIfKnownBool returns a value.Value that is set to the given bool if it is not null or unknown.
// Returns an unset value otherwise.
func SetIfKnownBool(v basetypes.BoolValue) *value.Value[bool] {
	if v.IsNull() || v.IsUnknown() {
		return value.Unset[bool]()
	}
	return value.Set(v.ValueBool())
}

// SetIfKnownInt returns a value.Value that is set to the given int if it is not null or unknown.
// Returns an unset value otherwise.
func SetIfKnownInt(v basetypes.Int32Value) *value.Value[int] {
	if v.IsNull() || v.IsUnknown() {
		return value.Unset[int]()
	}
	return value.Set(int(v.ValueInt32()))
}
