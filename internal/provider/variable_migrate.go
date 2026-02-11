package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"
)

func upgradeVariableResourceStateV0toV4(c *scalr.Client) func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	return func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		type variableModelV0 struct {
			Id          types.String `tfsdk:"id"`
			Key         types.String `tfsdk:"key"`
			Value       types.String `tfsdk:"value"`
			Category    types.String `tfsdk:"category"`
			HCL         types.Bool   `tfsdk:"hcl"`
			Sensitive   types.Bool   `tfsdk:"sensitive"`
			WorkspaceID types.String `tfsdk:"workspace_id"`
		}

		var dataV0 variableModelV0
		resp.Diagnostics.Append(req.State.Get(ctx, &dataV0)...)
		if resp.Diagnostics.HasError() {
			return
		}

		variable, err := c.Variables.Read(ctx, dataV0.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading variable", err.Error())
			return
		}

		data, diags := variableResourceModelFromAPI(ctx, variable, nil)
		if variable.Sensitive {
			data.Value = dataV0.Value
		}
		resp.Diagnostics.Append(diags...)

		diags = resp.State.Set(ctx, data)
		resp.Diagnostics.Append(diags...)
	}
}

func upgradeVariableResourceStateV1toV4(c *scalr.Client) func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	return func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		type variableModelV1 struct {
			Id            types.String `tfsdk:"id"`
			Key           types.String `tfsdk:"key"`
			Value         types.String `tfsdk:"value"`
			Category      types.String `tfsdk:"category"`
			HCL           types.Bool   `tfsdk:"hcl"`
			Sensitive     types.Bool   `tfsdk:"sensitive"`
			Final         types.Bool   `tfsdk:"final"`
			Force         types.Bool   `tfsdk:"force"`
			WorkspaceID   types.String `tfsdk:"workspace_id"`
			EnvironmentID types.String `tfsdk:"environment_id"`
			AccountID     types.String `tfsdk:"account_id"`
		}

		var dataV1 variableModelV1
		resp.Diagnostics.Append(req.State.Get(ctx, &dataV1)...)
		if resp.Diagnostics.HasError() {
			return
		}

		variable, err := c.Variables.Read(ctx, dataV1.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading variable", err.Error())
			return
		}

		data, diags := variableResourceModelFromAPI(ctx, variable, nil)
		if !dataV1.Force.IsNull() {
			data.Force = dataV1.Force
		}
		if variable.Sensitive {
			data.Value = dataV1.Value
		}
		resp.Diagnostics.Append(diags...)

		diags = resp.State.Set(ctx, data)
		resp.Diagnostics.Append(diags...)
	}
}

func upgradeVariableResourceStateV2toV4(c *scalr.Client) func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	return func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		type variableModelV2 struct {
			Id            types.String `tfsdk:"id"`
			Key           types.String `tfsdk:"key"`
			Value         types.String `tfsdk:"value"`
			Category      types.String `tfsdk:"category"`
			HCL           types.Bool   `tfsdk:"hcl"`
			Sensitive     types.Bool   `tfsdk:"sensitive"`
			Description   types.String `tfsdk:"description"`
			Final         types.Bool   `tfsdk:"final"`
			Force         types.Bool   `tfsdk:"force"`
			WorkspaceID   types.String `tfsdk:"workspace_id"`
			EnvironmentID types.String `tfsdk:"environment_id"`
			AccountID     types.String `tfsdk:"account_id"`
		}

		var dataV2 variableModelV2
		resp.Diagnostics.Append(req.State.Get(ctx, &dataV2)...)
		if resp.Diagnostics.HasError() {
			return
		}

		variable, err := c.Variables.Read(ctx, dataV2.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading variable", err.Error())
			return
		}

		data, diags := variableResourceModelFromAPI(ctx, variable, nil)
		if !dataV2.Force.IsNull() {
			data.Force = dataV2.Force
		}
		if variable.Sensitive {
			data.Value = dataV2.Value
		}
		resp.Diagnostics.Append(diags...)

		diags = resp.State.Set(ctx, data)
		resp.Diagnostics.Append(diags...)
	}
}

func upgradeVariableResourceStateV3toV4(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	type variableModelV3 struct {
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

	var dataV3 variableModelV3
	resp.Diagnostics.Append(req.State.Get(ctx, &dataV3)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// V3 -> V4: Add value_wo and value_wo_version fields (initialized to null)
	dataV4 := variableResourceModel{
		Id:             dataV3.Id,
		Key:            dataV3.Key,
		Value:          dataV3.Value,
		ValueWO:        types.StringNull(),
		ValueWOVersion: types.Int64Null(),
		ReadableValue:  dataV3.ReadableValue,
		Category:       dataV3.Category,
		HCL:            dataV3.HCL,
		Sensitive:      dataV3.Sensitive,
		Description:    dataV3.Description,
		Final:          dataV3.Final,
		Force:          dataV3.Force,
		WorkspaceID:    dataV3.WorkspaceID,
		EnvironmentID:  dataV3.EnvironmentID,
		AccountID:      dataV3.AccountID,
		UpdatedAt:      dataV3.UpdatedAt,
		UpdatedByEmail: dataV3.UpdatedByEmail,
		UpdatedBy:      dataV3.UpdatedBy,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, dataV4)...)
}
