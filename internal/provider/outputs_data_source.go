package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/ops/workspace"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

var (
	_ datasource.DataSource              = &workspaceOutputsDataSource{}
	_ datasource.DataSourceWithConfigure = &workspaceOutputsDataSource{}
)

func outputsDataSource() datasource.DataSource {
	return &workspaceOutputsDataSource{}
}

// workspaceOutputsDataSource defines the data source implementation.
type workspaceOutputsDataSource struct {
	framework.DataSourceWithScalrClient
}

type outputsModel struct {
	ID                 types.String  `tfsdk:"id"`
	Environment        types.String  `tfsdk:"environment"`
	Workspace          types.String  `tfsdk:"workspace"`
	Values             types.Dynamic `tfsdk:"values"`
	NonSensitiveValues types.Dynamic `tfsdk:"nonsensitive_values"`
}

func (d *workspaceOutputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_outputs"
}

func (d *workspaceOutputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the outputs of a Scalr workspace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the workspace.",
				Computed:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The name of the environment the workspace belongs to.",
				Required:            true,
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "The name of the workspace.",
				Required:            true,
			},
			"values": schema.DynamicAttribute{
				MarkdownDescription: "A map of all workspace output values.",
				Computed:            true,
				Sensitive:           true,
			},
			"nonsensitive_values": schema.DynamicAttribute{
				MarkdownDescription: "A map of non-sensitive workspace output values.",
				Computed:            true,
			},
		},
	}
}

func (d *workspaceOutputsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg outputsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaces, err := d.ClientV2.Workspace.GetWorkspaces(ctx, &workspace.GetWorkspacesOptions{
		Filter: map[string]string{
			"name":              cfg.Workspace.ValueString(),
			"environment][name": cfg.Environment.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing workspaces", err.Error())
		return
	}
	if len(workspaces) == 0 {
		resp.Diagnostics.AddError(
			"Workspace not found",
			fmt.Sprintf(
				"No workspace %q found in environment %q.",
				cfg.Workspace.ValueString(),
				cfg.Environment.ValueString(),
			),
		)
		return
	} else if len(workspaces) > 1 {
		resp.Diagnostics.AddError(
			"Multiple workspaces found",
			fmt.Sprintf(
				"Multiple workspaces %q found in environment %q.",
				cfg.Workspace.ValueString(),
				cfg.Environment.ValueString(),
			),
		)
	}

	wsID := workspaces[0].ID
	cfg.ID = types.StringValue(wsID)

	outputsJSON, err := d.ClientV2.Workspace.GetWorkspaceOutputs(ctx, wsID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workspace outputs", err.Error())
		return
	}

	var outputsResp struct {
		Data []struct {
			Name      string          `json:"name"`
			Value     json.RawMessage `json:"value"`
			Sensitive bool            `json:"sensitive"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(outputsJSON), &outputsResp); err != nil {
		resp.Diagnostics.AddError("Error parsing workspace outputs", err.Error())
		return
	}

	allAttrTypes := make(map[string]attr.Type)
	allAttrValues := make(map[string]attr.Value)
	nsAttrTypes := make(map[string]attr.Type)
	nsAttrValues := make(map[string]attr.Value)

	for _, output := range outputsResp.Data {
		attrType, attrValue, diags := jsonRawToAttrValue(output.Value)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		allAttrTypes[output.Name] = attrType
		allAttrValues[output.Name] = attrValue
		if !output.Sensitive {
			nsAttrTypes[output.Name] = attrType
			nsAttrValues[output.Name] = attrValue
		}
	}

	allObj, diags := types.ObjectValue(allAttrTypes, allAttrValues)
	resp.Diagnostics.Append(diags...)
	cfg.Values = types.DynamicValue(allObj)

	nsObj, diags := types.ObjectValue(nsAttrTypes, nsAttrValues)
	resp.Diagnostics.Append(diags...)
	cfg.NonSensitiveValues = types.DynamicValue(nsObj)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

func jsonRawToAttrValue(raw json.RawMessage) (attr.Type, attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(raw) == 0 || string(raw) == "null" {
		return types.StringType, types.StringNull(), diags
	}

	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return types.StringType, types.StringValue(string(raw)), diags
	}

	attrType, err := inferAttrType(v)
	if err != nil {
		diags.AddError("Error inferring attribute type", err.Error())
		return types.StringType, types.StringNull(), diags
	}

	attrValue, ds := convertToValue(v, attrType)
	diags.Append(ds...)

	return attrType, attrValue, diags
}

func inferAttrType(raw interface{}) (attr.Type, error) {
	if raw == nil {
		return types.StringType, nil
	}

	switch v := raw.(type) {
	case bool:
		return types.BoolType, nil
	case int, int8, int16, int32, int64, float32, float64:
		return types.NumberType, nil
	case string:
		return types.StringType, nil
	case []interface{}:
		if len(v) == 0 {
			return types.ListType{ElemType: types.DynamicType}, nil
		}

		elemTypes := make([]attr.Type, len(v))
		for i, elem := range v {
			t, err := inferAttrType(elem)
			if err != nil {
				return nil, err
			}
			elemTypes[i] = t
		}

		for i := 1; i < len(elemTypes); i++ {
			if !reflect.DeepEqual(elemTypes[0], elemTypes[i]) {
				return types.TupleType{ElemTypes: elemTypes}, nil
			}
		}
		return types.ListType{ElemType: elemTypes[0]}, nil
	case map[string]interface{}:
		attrTypes := make(map[string]attr.Type)
		for key, val := range v {
			inferred, err := inferAttrType(val)
			if err != nil {
				return nil, fmt.Errorf("error inferring type for key %q: %w", key, err)
			}
			attrTypes[key] = inferred
		}
		return types.ObjectType{AttrTypes: attrTypes}, nil

	default:
		return nil, fmt.Errorf("unsupported type %T", raw)
	}
}

func convertToValue(raw interface{}, t attr.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == nil {
		return types.StringNull(), diags
	}

	if t == types.BoolType {
		b, ok := raw.(bool)
		if !ok {
			diags.AddError("Conversion Error", "expected bool")
			return types.BoolNull(), diags
		}
		return types.BoolValue(b), diags
	}

	if t == types.NumberType {
		n, ok := raw.(float64)
		if !ok {
			diags.AddError("Conversion Error", "expected number")
			return types.NumberNull(), diags
		}
		return types.NumberValue(big.NewFloat(n)), diags
	}

	if t == types.StringType {
		s, ok := raw.(string)
		if !ok {
			diags.AddError("Conversion Error", "expected string")
			return types.StringNull(), diags
		}
		return types.StringValue(s), diags
	}

	if tt, ok := t.(types.ListType); ok {
		slice, ok := raw.([]interface{})
		if !ok {
			diags.AddError("Conversion Error", "expected slice for ListType")
			return types.ListNull(tt.ElemType), diags
		}

		var elems []attr.Value
		for _, elem := range slice {
			v, ds := convertToValue(elem, tt.ElemType)
			diags.Append(ds...)
			elems = append(elems, v)
		}
		return types.ListValue(tt.ElemType, elems)
	}

	if tt, ok := t.(types.TupleType); ok {
		slice, ok := raw.([]interface{})
		if !ok {
			diags.AddError("Conversion Error", "expected slice for TupleType")
			return types.TupleNull(tt.ElemTypes), diags
		}
		if len(slice) != len(tt.ElemTypes) {
			diags.AddError("Conversion Error", "tuple length mismatch")
			return types.TupleNull(tt.ElemTypes), diags
		}

		var elems []attr.Value
		for i, elem := range slice {
			v, ds := convertToValue(elem, tt.ElemTypes[i])
			diags.Append(ds...)
			elems = append(elems, v)
		}
		return types.TupleValue(tt.ElemTypes, elems)
	}

	if tt, ok := t.(types.ObjectType); ok {
		m, ok := raw.(map[string]interface{})
		if !ok {
			diags.AddError("Conversion Error", "expected map for ObjectType")
			return types.ObjectNull(tt.AttrTypes), diags
		}

		objValues := make(map[string]attr.Value)
		for key, expectedType := range tt.AttrTypes {
			value := m[key]
			v, ds := convertToValue(value, expectedType)
			diags.Append(ds...)
			if ds.HasError() {
				return types.ObjectNull(tt.AttrTypes), diags
			}
			objValues[key] = v
		}
		return types.ObjectValue(tt.AttrTypes, objValues)
	}

	if tt, ok := t.(types.MapType); ok {
		m, ok := raw.(map[string]interface{})
		if !ok {
			diags.AddError("Conversion Error", "expected map for MapType")
			return types.MapValue(tt.ElemType, nil)
		}

		mapValues := make(map[string]attr.Value)
		for key, value := range m {
			v, ds := convertToValue(value, tt.ElemType)
			diags.Append(ds...)
			if ds.HasError() {
				return types.MapValue(tt.ElemType, nil)
			}
			mapValues[key] = v
		}
		return types.MapValue(tt.ElemType, mapValues)
	}

	diags.AddError("Conversion Error", fmt.Sprintf("unsupported type %T", t))
	return nil, diags
}
