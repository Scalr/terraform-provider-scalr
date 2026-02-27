package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestJsonRawToAttrValue_null(t *testing.T) {
	attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage("null"))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if attrType != types.StringType {
		t.Errorf("expected StringType, got %T", attrType)
	}
	if attrValue != types.StringNull() {
		t.Errorf("expected StringNull, got %v", attrValue)
	}
}

func TestJsonRawToAttrValue_empty(t *testing.T) {
	attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(""))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if attrType != types.StringType {
		t.Errorf("expected StringType, got %T", attrType)
	}
	if attrValue != types.StringNull() {
		t.Errorf("expected StringNull, got %v", attrValue)
	}
}

func TestJsonRawToAttrValue_bool(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want bool
	}{
		{"true", true},
		{"false", false},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(tc.raw))
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			if attrType != types.BoolType {
				t.Errorf("expected BoolType, got %T", attrType)
			}
			if attrValue != types.BoolValue(tc.want) {
				t.Errorf("expected BoolValue(%v), got %v", tc.want, attrValue)
			}
		})
	}
}

func TestJsonRawToAttrValue_number(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want float64
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"-7", -7},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(tc.raw))
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			if attrType != types.NumberType {
				t.Errorf("expected NumberType, got %T", attrType)
			}
			numVal, ok := attrValue.(types.Number)
			if !ok {
				t.Fatalf("expected types.Number, got %T", attrValue)
			}
			got, _ := numVal.ValueBigFloat().Float64()
			want, _ := big.NewFloat(tc.want).Float64()
			if got != want {
				t.Errorf("expected %v, got %v", want, got)
			}
		})
	}
}

func TestJsonRawToAttrValue_string(t *testing.T) {
	attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(`"hello world"`))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if attrType != types.StringType {
		t.Errorf("expected StringType, got %T", attrType)
	}
	if attrValue != types.StringValue("hello world") {
		t.Errorf("expected StringValue(\"hello world\"), got %v", attrValue)
	}
}

func TestJsonRawToAttrValue_homogeneousList(t *testing.T) {
	attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(`[1, 2, 3]`))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	wantType := types.ListType{ElemType: types.NumberType}
	if !reflect.DeepEqual(attrType, wantType) {
		t.Errorf("expected %v, got %v", wantType, attrType)
	}
	listVal, ok := attrValue.(types.List)
	if !ok {
		t.Fatalf("expected types.List, got %T", attrValue)
	}
	if listVal.IsNull() || listVal.IsUnknown() {
		t.Error("expected non-null, non-unknown list")
	}
	if listVal.ElementType(context.TODO()) != types.NumberType {
		t.Errorf("unexpected element type: %T", listVal.ElementType(context.TODO()))
	}
}

func TestJsonRawToAttrValue_emptyList(t *testing.T) {
	attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(`[]`))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	wantType := types.ListType{ElemType: types.DynamicType}
	if !reflect.DeepEqual(attrType, wantType) {
		t.Errorf("expected %v, got %v", wantType, attrType)
	}
	if _, ok := attrValue.(types.List); !ok {
		t.Fatalf("expected types.List, got %T", attrValue)
	}
}

func TestJsonRawToAttrValue_tuple(t *testing.T) {
	attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(`[1, "two", true]`))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	wantType := types.TupleType{ElemTypes: []attr.Type{types.NumberType, types.StringType, types.BoolType}}
	if !reflect.DeepEqual(attrType, wantType) {
		t.Errorf("expected %v, got %v", wantType, attrType)
	}
	if _, ok := attrValue.(types.Tuple); !ok {
		t.Fatalf("expected types.Tuple, got %T", attrValue)
	}
}

func TestJsonRawToAttrValue_object(t *testing.T) {
	attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(`{"name": "scalr", "count": 2, "active": true}`))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	wantType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":   types.StringType,
		"count":  types.NumberType,
		"active": types.BoolType,
	}}
	if !reflect.DeepEqual(attrType, wantType) {
		t.Errorf("expected %v, got %v", wantType, attrType)
	}
	obj, ok := attrValue.(types.Object)
	if !ok {
		t.Fatalf("expected types.Object, got %T", attrValue)
	}
	if obj.IsNull() || obj.IsUnknown() {
		t.Error("expected non-null, non-unknown object")
	}
}

func TestJsonRawToAttrValue_nestedObject(t *testing.T) {
	attrType, attrValue, diags := jsonRawToAttrValue(json.RawMessage(`{"outer": {"inner": "value"}}`))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	outerObjType, ok := attrType.(types.ObjectType)
	if !ok {
		t.Fatalf("expected ObjectType, got %T", attrType)
	}
	innerType, exists := outerObjType.AttrTypes["outer"]
	if !exists {
		t.Fatal("expected 'outer' key in ObjectType")
	}
	innerObjType, ok := innerType.(types.ObjectType)
	if !ok {
		t.Fatalf("expected inner ObjectType, got %T", innerType)
	}
	if innerObjType.AttrTypes["inner"] != types.StringType {
		t.Errorf("expected inner.inner = StringType, got %T", innerObjType.AttrTypes["inner"])
	}
	if _, ok := attrValue.(types.Object); !ok {
		t.Fatalf("expected types.Object, got %T", attrValue)
	}
}

func TestAccScalrWorkspaceOutputsDataSource_basic(t *testing.T) {
	rInt := GetRandomInteger()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: protoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config:      testAccScalrWorkspaceOutputsDataSourceNotFoundConfig(rInt),
				ExpectError: regexp.MustCompile("Workspace not found"),
				PlanOnly:    true,
			},
			{
				Config: testAccScalrWorkspaceOutputsDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scalr_outputs.test", "id"),
				),
			},
		},
	})
}

func testAccScalrWorkspaceOutputsDataSourceNotFoundConfig(rInt int) string {
	return fmt.Sprintf(`
data scalr_outputs test {
  environment = "nonexistent-env-%[1]d"
  workspace   = "nonexistent-ws-%[1]d"
}`, rInt)
}

func testAccScalrWorkspaceOutputsDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource scalr_environment test {
  name       = "test-env-%[1]d"
  account_id = "%s"
}

resource scalr_workspace test {
  name           = "workspace-test-%[1]d"
  environment_id = scalr_environment.test.id
}

data scalr_outputs test {
  environment = scalr_environment.test.name
  workspace   = scalr_workspace.test.name
}`, rInt, defaultAccount)
}
