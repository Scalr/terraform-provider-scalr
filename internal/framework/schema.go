package framework

import (
	"context"
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func FlattenRelationshipIDsSet[T any](ctx context.Context, apiValue []*T, idFunc func(*T) string, prior *types.Set) (
	types.Set,
	diag.Diagnostics,
) {
	if len(apiValue) > 0 {
		ids := make([]string, 0, len(apiValue))
		for _, v := range apiValue {
			if v == nil {
				continue
			}
			ids = append(ids, idFunc(v))
		}
		if len(ids) > 0 {
			return types.SetValueFrom(ctx, types.StringType, ids)
		}
	}

	if prior != nil && prior.IsNull() {
		return types.SetNull(types.StringType), nil
	}

	// preserve explicit empty set
	return types.SetValueFrom(ctx, types.StringType, []string{})
}

func FlattenRelationshipIDsList[T any](ctx context.Context, apiValue []*T, idFunc func(*T) string, prior *types.List) (
	types.List,
	diag.Diagnostics,
) {
	if len(apiValue) > 0 {
		ids := make([]string, 0, len(apiValue))
		for _, v := range apiValue {
			if v == nil {
				continue
			}
			ids = append(ids, idFunc(v))
		}
		if len(ids) > 0 {
			return types.ListValueFrom(ctx, types.StringType, ids)
		}
	}

	if prior != nil && prior.IsNull() {
		return types.ListNull(types.StringType), nil
	}

	// preserve explicit empty list
	return types.ListValueFrom(ctx, types.StringType, []string{})
}

func ExpandRelationshipIDsSet[T any](ctx context.Context, v types.Set, newFunc func(string) T) (
	[]T,
	diag.Diagnostics,
) {
	var (
		ids   []string
		diags diag.Diagnostics
	)

	diags.Append(v.ElementsAs(ctx, &ids, false)...)
	if diags.HasError() {
		return nil, diags
	}

	rels := make([]T, len(ids))
	for i, id := range ids {
		rels[i] = newFunc(id)
	}
	return rels, diags
}

func ExpandRelationshipIDsList[T any](ctx context.Context, v types.List, newFunc func(string) T) (
	[]T,
	diag.Diagnostics,
) {
	var (
		ids   []string
		diags diag.Diagnostics
	)

	diags.Append(v.ElementsAs(ctx, &ids, false)...)
	if diags.HasError() {
		return nil, diags
	}

	rels := make([]T, len(ids))
	for i, id := range ids {
		rels[i] = newFunc(id)
	}
	return rels, diags
}
