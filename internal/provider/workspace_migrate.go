package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
	"github.com/scalr/go-scalr/v2/scalr/ops/workspace"
)

func upgradeWorkspaceResourceStateV0toV4(c *scalrV2.Client) func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	return func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		type workspaceModelV0 struct {
			Id               types.String `tfsdk:"id"`
			AutoApply        types.Bool   `tfsdk:"auto_apply"`
			CreatedBy        types.List   `tfsdk:"created_by"`
			ExternalID       types.String `tfsdk:"external_id"`
			Name             types.String `tfsdk:"name"`
			Operations       types.Bool   `tfsdk:"operations"`
			Organization     types.String `tfsdk:"organization"`
			SSHKeyID         types.String `tfsdk:"ssh_key_id"`
			TerraformVersion types.String `tfsdk:"terraform_version"`
			VCSRepo          types.List   `tfsdk:"vcs_repo"`
			WorkingDirectory types.String `tfsdk:"working_directory"`
		}

		var dataV0 workspaceModelV0
		resp.Diagnostics.Append(req.State.Get(ctx, &dataV0)...)
		if resp.Diagnostics.HasError() {
			return
		}

		ws, err := c.Workspace.GetWorkspace(ctx, dataV0.Id.ValueString(), &workspace.GetWorkspaceOptions{
			Include: []string{"created-by"},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error reading workspace", err.Error())
			return
		}

		pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, c, dataV0.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
			return
		}

		stateConsumers, err := getRemoteStateConsumers(ctx, c, dataV0.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
			return
		}

		data, diags := workspaceResourceModelFromAPI(ctx, ws, pcfgLinks, stateConsumers, nil)
		resp.Diagnostics.Append(diags...)

		diags = resp.State.Set(ctx, data)
		resp.Diagnostics.Append(diags...)
	}
}

func upgradeWorkspaceResourceStateV1toV4(c *scalrV2.Client) func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	return func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		type workspaceModelV1 struct {
			Id               types.String `tfsdk:"id"`
			AutoApply        types.Bool   `tfsdk:"auto_apply"`
			CreatedBy        types.List   `tfsdk:"created_by"`
			EnvironmentID    types.String `tfsdk:"environment_id"`
			Name             types.String `tfsdk:"name"`
			Operations       types.Bool   `tfsdk:"operations"`
			QueueAllRuns     types.Bool   `tfsdk:"queue_all_runs"`
			TerraformVersion types.String `tfsdk:"terraform_version"`
			VCSRepo          types.List   `tfsdk:"vcs_repo"`
			WorkingDirectory types.String `tfsdk:"working_directory"`
		}

		var dataV1 workspaceModelV1
		resp.Diagnostics.Append(req.State.Get(ctx, &dataV1)...)
		if resp.Diagnostics.HasError() {
			return
		}

		ws, err := c.Workspace.GetWorkspace(ctx, dataV1.Id.ValueString(), &workspace.GetWorkspaceOptions{
			Include: []string{"created-by"},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error reading workspace", err.Error())
			return
		}

		pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, c, dataV1.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
			return
		}

		stateConsumers, err := getRemoteStateConsumers(ctx, c, dataV1.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
			return
		}

		data, diags := workspaceResourceModelFromAPI(ctx, ws, pcfgLinks, stateConsumers, nil)
		resp.Diagnostics.Append(diags...)

		diags = resp.State.Set(ctx, data)
		resp.Diagnostics.Append(diags...)
	}
}

func upgradeWorkspaceResourceStateV2toV4(c *scalrV2.Client) func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	return func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		type workspaceModelV2 struct {
			Id               types.String `tfsdk:"id"`
			AutoApply        types.Bool   `tfsdk:"auto_apply"`
			CreatedBy        types.List   `tfsdk:"created_by"`
			EnvironmentID    types.String `tfsdk:"environment_id"`
			Name             types.String `tfsdk:"name"`
			Operations       types.Bool   `tfsdk:"operations"`
			TerraformVersion types.String `tfsdk:"terraform_version"`
			VCSProviderID    types.String `tfsdk:"vcs_provider_id"`
			VCSRepo          types.List   `tfsdk:"vcs_repo"`
			WorkingDirectory types.String `tfsdk:"working_directory"`
		}

		var dataV2 workspaceModelV2
		resp.Diagnostics.Append(req.State.Get(ctx, &dataV2)...)
		if resp.Diagnostics.HasError() {
			return
		}

		ws, err := c.Workspace.GetWorkspace(ctx, dataV2.Id.ValueString(), &workspace.GetWorkspaceOptions{
			Include: []string{"created-by"},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error reading workspace", err.Error())
			return
		}

		pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, c, dataV2.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
			return
		}

		stateConsumers, err := getRemoteStateConsumers(ctx, c, dataV2.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
			return
		}

		data, diags := workspaceResourceModelFromAPI(ctx, ws, pcfgLinks, stateConsumers, nil)
		resp.Diagnostics.Append(diags...)

		diags = resp.State.Set(ctx, data)
		resp.Diagnostics.Append(diags...)
	}
}

func upgradeWorkspaceResourceStateV3toV4(c *scalrV2.Client) func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	return func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		type workspaceModelV3 struct {
			Id                    types.String `tfsdk:"id"`
			AgentPoolID           types.String `tfsdk:"agent_pool_id"`
			AutoApply             types.Bool   `tfsdk:"auto_apply"`
			CreatedBy             types.List   `tfsdk:"created_by"`
			EnvironmentID         types.String `tfsdk:"environment_id"`
			HasResources          types.Bool   `tfsdk:"has_resources"`
			Hooks                 types.List   `tfsdk:"hooks"`
			ModuleVersionID       types.String `tfsdk:"module_version_id"`
			Name                  types.String `tfsdk:"name"`
			Operations            types.Bool   `tfsdk:"operations"`
			ProviderConfiguration types.Set    `tfsdk:"provider_configuration"`
			RunOperationTimeout   types.Int32  `tfsdk:"run_operation_timeout"`
			TerraformVersion      types.String `tfsdk:"terraform_version"`
			VCSProviderID         types.String `tfsdk:"vcs_provider_id"`
			VCSRepo               types.List   `tfsdk:"vcs_repo"`
			VarFiles              types.List   `tfsdk:"var_files"`
			WorkingDirectory      types.String `tfsdk:"working_directory"`
		}

		var dataV3 workspaceModelV3
		resp.Diagnostics.Append(req.State.Get(ctx, &dataV3)...)
		if resp.Diagnostics.HasError() {
			return
		}

		ws, err := c.Workspace.GetWorkspace(ctx, dataV3.Id.ValueString(), &workspace.GetWorkspaceOptions{
			Include: []string{"created-by"},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error reading workspace", err.Error())
			return
		}

		pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, c, dataV3.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
			return
		}

		stateConsumers, err := getRemoteStateConsumers(ctx, c, dataV3.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
			return
		}

		data, diags := workspaceResourceModelFromAPI(ctx, ws, pcfgLinks, stateConsumers, nil)
		resp.Diagnostics.Append(diags...)

		diags = resp.State.Set(ctx, data)
		resp.Diagnostics.Append(diags...)
	}
}
