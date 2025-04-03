package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"
)

type variableResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
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
	UpdatedAt      types.String `tfsdk:"updated_at"`
	UpdatedByEmail types.String `tfsdk:"updated_by_email"`
	UpdatedBy      types.List   `tfsdk:"updated_by"`
}

func variableResourceModelFromAPI(ctx context.Context, v *scalr.Variable, existing *variableResourceModel) (*variableResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &variableResourceModel{
		Id:             types.StringValue(v.ID),
		Key:            types.StringValue(v.Key),
		Value:          types.StringNull(),
		ReadableValue:  types.StringNull(),
		Category:       types.StringValue(string(v.Category)),
		HCL:            types.BoolValue(v.HCL),
		Sensitive:      types.BoolValue(v.Sensitive),
		Description:    types.StringValue(v.Description),
		Final:          types.BoolValue(v.Final),
		Force:          types.BoolValue(false),
		WorkspaceID:    types.StringNull(),
		EnvironmentID:  types.StringNull(),
		AccountID:      types.StringNull(),
		UpdatedAt:      types.StringNull(),
		UpdatedByEmail: types.StringValue(v.UpdatedByEmail),
		UpdatedBy:      types.ListNull(userElementType),
	}

	if existing != nil && !existing.Force.IsUnknown() && !existing.Force.IsNull() {
		model.Force = existing.Force
	}

	if v.Workspace != nil {
		model.WorkspaceID = types.StringValue(v.Workspace.ID)
	}

	if v.Environment != nil {
		model.EnvironmentID = types.StringValue(v.Environment.ID)
	}

	if v.Account != nil {
		model.AccountID = types.StringValue(v.Account.ID)
	}

	if v.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(v.UpdatedAt.Format(time.RFC3339))
	}

	if v.UpdatedBy != nil {
		updatedBy := []userModel{*userModelFromAPI(v.UpdatedBy)}
		updatedByValue, d := types.ListValueFrom(ctx, userElementType, updatedBy)
		diags.Append(d...)
		model.UpdatedBy = updatedByValue
	}

	// Only set the value if the variable is not sensitive, as otherwise it will be empty.
	if !v.Sensitive {
		model.Value = types.StringValue(v.Value)
		model.ReadableValue = model.Value
	} else if existing != nil {
		model.Value = existing.Value
	}

	return model, diags
}
