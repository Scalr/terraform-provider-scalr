package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/schemas"
)

type variableResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	ValueWO        types.String `tfsdk:"value_wo"`
	ValueWOVersion types.Int64  `tfsdk:"value_wo_version"`
	ReadableValue  types.String `tfsdk:"readable_value"`
	Category       types.String `tfsdk:"category"`
	HCL            types.Bool   `tfsdk:"hcl"`
	Sensitive      types.Bool   `tfsdk:"sensitive"`
	Description    types.String `tfsdk:"description"`
	Final          types.Bool   `tfsdk:"final"`
	Force          types.Bool   `tfsdk:"force"`
	WorkspaceID    types.String `tfsdk:"workspace_id"`
	EnvironmentID  types.String `tfsdk:"environment_id"`
	AccountID      types.String `tfsdk:"account_id"`
	VarSetID       types.String `tfsdk:"var_set_id"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	UpdatedByEmail types.String `tfsdk:"updated_by_email"`
	UpdatedBy      types.List   `tfsdk:"updated_by"`
}

func variableResourceModelFromAPI(
	ctx context.Context,
	v *schemas.Variable,
	existing *variableResourceModel,
	isWriteOnly bool,
) (*variableResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &variableResourceModel{
		Id:             types.StringValue(v.ID),
		Key:            types.StringValue(v.Attributes.Key),
		Value:          types.StringNull(),
		ValueWO:        types.StringNull(),
		ValueWOVersion: types.Int64Null(),
		ReadableValue:  types.StringNull(),
		Category:       types.StringValue(string(v.Attributes.Category)),
		HCL:            types.BoolValue(v.Attributes.Hcl),
		Sensitive:      types.BoolValue(v.Attributes.Sensitive),
		Description:    types.StringPointerValue(v.Attributes.Description),
		Final:          types.BoolValue(v.Attributes.Final),
		Force:          types.BoolValue(false),
		WorkspaceID:    types.StringNull(),
		EnvironmentID:  types.StringNull(),
		AccountID:      types.StringNull(),
		VarSetID:       types.StringNull(), // keep it null for classic variables
		UpdatedAt:      types.StringNull(),
		UpdatedByEmail: types.StringPointerValue(v.Attributes.UpdatedByEmail),
		UpdatedBy:      types.ListNull(userElementType),
	}

	if existing != nil && !existing.Force.IsUnknown() && !existing.Force.IsNull() {
		model.Force = existing.Force
	}

	// Preserve value_wo_version (it does not come from the API)
	if existing != nil && !existing.ValueWOVersion.IsUnknown() && !existing.ValueWOVersion.IsNull() {
		model.ValueWOVersion = existing.ValueWOVersion
	}

	if v.Relationships.Workspace != nil {
		model.WorkspaceID = types.StringValue(v.Relationships.Workspace.ID)
	}

	if v.Relationships.Environment != nil {
		model.EnvironmentID = types.StringValue(v.Relationships.Environment.ID)
	}

	if v.Relationships.Account != nil {
		model.AccountID = types.StringValue(v.Relationships.Account.ID)
	}

	if v.Attributes.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(v.Attributes.UpdatedAt.Format(time.RFC3339))
	}

	if v.Relationships.UpdatedBy != nil {
		updatedBy := []userModel{*userModelFromAPIv2(v.Relationships.UpdatedBy)}
		updatedByValue, d := types.ListValueFrom(ctx, userElementType, updatedBy)
		diags.Append(d...)
		model.UpdatedBy = updatedByValue
	}

	// Only set the value if the variable is not sensitive, as otherwise it will be empty.
	if !v.Attributes.Sensitive {
		model.Value = types.StringPointerValue(v.Attributes.Value)
		model.ReadableValue = model.Value
	} else if existing != nil {
		model.Value = existing.Value
	}
	// Unset value and readable_value if write-only value was used.
	if isWriteOnly {
		model.Value = types.StringValue("") // it has a default value of empty string in the schema
		model.ReadableValue = types.StringNull()
	}

	return model, diags
}

// variableSetVariableResourceModelFromAPI duplicates variableResourceModelFromAPI almost entirely,
// with adjustments for a variable set variable.
func variableSetVariableResourceModelFromAPI(
	_ context.Context,
	v *schemas.VariableSetVariable,
	existing *variableResourceModel,
	isWriteOnly bool,
) (*variableResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &variableResourceModel{
		Id:             types.StringValue(v.ID),
		Key:            types.StringValue(v.Attributes.Key),
		Value:          types.StringNull(),
		ValueWO:        types.StringNull(),
		ValueWOVersion: types.Int64Null(),
		ReadableValue:  types.StringNull(),
		Category:       types.StringValue(string(v.Attributes.Category)),
		HCL:            types.BoolValue(v.Attributes.Hcl),
		Sensitive:      types.BoolValue(v.Attributes.Sensitive),
		Description:    types.StringPointerValue(v.Attributes.Description),
		Final:          types.BoolValue(v.Attributes.Final),
		Force:          types.BoolValue(false),
		WorkspaceID:    types.StringNull(), // keep it null for variable set variables
		EnvironmentID:  types.StringNull(), // keep it null for variable set variables
		AccountID:      types.StringNull(),
		VarSetID:       types.StringNull(),
		UpdatedAt:      types.StringValue(v.Attributes.UpdatedAt.Format(time.RFC3339)),
		UpdatedByEmail: types.StringPointerValue(v.Attributes.UpdatedByEmail),
		UpdatedBy:      types.ListNull(userElementType),
	}

	if existing != nil && !existing.Force.IsUnknown() && !existing.Force.IsNull() {
		model.Force = existing.Force
	}

	// Preserve value_wo_version (it does not come from the API)
	if existing != nil && !existing.ValueWOVersion.IsUnknown() && !existing.ValueWOVersion.IsNull() {
		model.ValueWOVersion = existing.ValueWOVersion
	}

	if v.Relationships.VarSet != nil {
		model.VarSetID = types.StringValue(v.Relationships.VarSet.ID)
	}

	if v.Relationships.Account != nil {
		model.AccountID = types.StringValue(v.Relationships.Account.ID)
	}

	// Only set the value if the variable is not sensitive, as otherwise it will be empty.
	if !v.Attributes.Sensitive {
		model.Value = types.StringPointerValue(v.Attributes.Value)
		model.ReadableValue = model.Value
	} else if existing != nil {
		model.Value = existing.Value
	}
	// Unset value and readable_value if write-only value was used.
	if isWriteOnly {
		model.Value = types.StringValue("") // it has a default value of empty string in the schema
		model.ReadableValue = types.StringNull()
	}

	return model, diags
}
