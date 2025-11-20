package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/schemas"
)

var (
	vcsRepoElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"branch":             types.StringType,
			"dry_runs_enabled":   types.BoolType,
			"identifier":         types.StringType,
			"ingress_submodules": types.BoolType,
			"path":               types.StringType,
			"trigger_patterns":   types.StringType,
			"trigger_prefixes":   types.ListType{ElemType: types.StringType},
			"version_constraint": types.StringType,
		},
	}
	terragruntElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"version":                       types.StringType,
			"use_run_all":                   types.BoolType,
			"include_external_dependencies": types.BoolType,
		},
	}

	providerConfigurationElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":    types.StringType,
			"alias": types.StringType,
		},
	}
	hooksElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"pre_init":   types.StringType,
			"pre_plan":   types.StringType,
			"post_plan":  types.StringType,
			"pre_apply":  types.StringType,
			"post_apply": types.StringType,
		},
	}
)

type workspaceResourceModel struct {
	Id                        types.String `tfsdk:"id"`
	AgentPoolID               types.String `tfsdk:"agent_pool_id"`
	AutoApply                 types.Bool   `tfsdk:"auto_apply"`
	AutoQueueRuns             types.String `tfsdk:"auto_queue_runs"`
	CreatedBy                 types.List   `tfsdk:"created_by"`
	DeletionProtectionEnabled types.Bool   `tfsdk:"deletion_protection_enabled"`
	EnvironmentID             types.String `tfsdk:"environment_id"`
	ExecutionMode             types.String `tfsdk:"execution_mode"`
	ForceLatestRun            types.Bool   `tfsdk:"force_latest_run"`
	HasResources              types.Bool   `tfsdk:"has_resources"`
	Hooks                     types.List   `tfsdk:"hooks"`
	IaCPlatform               types.String `tfsdk:"iac_platform"`
	ModuleVersionID           types.String `tfsdk:"module_version_id"`
	Name                      types.String `tfsdk:"name"`
	Operations                types.Bool   `tfsdk:"operations"`
	ProviderConfiguration     types.Set    `tfsdk:"provider_configuration"`
	RemoteBackend             types.Bool   `tfsdk:"remote_backend"`
	RemoteStateConsumers      types.Set    `tfsdk:"remote_state_consumers"`
	RunOperationTimeout       types.Int32  `tfsdk:"run_operation_timeout"`
	SSHKeyID                  types.String `tfsdk:"ssh_key_id"`
	TagIDs                    types.Set    `tfsdk:"tag_ids"`
	TerraformVersion          types.String `tfsdk:"terraform_version"`
	Terragrunt                types.List   `tfsdk:"terragrunt"`
	Type                      types.String `tfsdk:"type"`
	VarFiles                  types.List   `tfsdk:"var_files"`
	VCSProviderID             types.String `tfsdk:"vcs_provider_id"`
	VCSRepo                   types.List   `tfsdk:"vcs_repo"`
	WorkingDirectory          types.String `tfsdk:"working_directory"`
}

type terragruntModel struct {
	Version                     types.String `tfsdk:"version"`
	UseRunAll                   types.Bool   `tfsdk:"use_run_all"`
	IncludeExternalDependencies types.Bool   `tfsdk:"include_external_dependencies"`
}

type vcsRepoModel struct {
	Branch            types.String `tfsdk:"branch"`
	DryRunsEnabled    types.Bool   `tfsdk:"dry_runs_enabled"`
	Identifier        types.String `tfsdk:"identifier"`
	IngressSubmodules types.Bool   `tfsdk:"ingress_submodules"`
	Path              types.String `tfsdk:"path"`
	TriggerPatterns   types.String `tfsdk:"trigger_patterns"`
	TriggerPrefixes   types.List   `tfsdk:"trigger_prefixes"`
	VersionConstraint types.String `tfsdk:"version_constraint"`
}

type providerConfigurationModel struct {
	ID    types.String `tfsdk:"id"`
	Alias types.String `tfsdk:"alias"`
}

type hooksModel struct {
	PreInit   types.String `tfsdk:"pre_init"`
	PrePlan   types.String `tfsdk:"pre_plan"`
	PostPlan  types.String `tfsdk:"post_plan"`
	PreApply  types.String `tfsdk:"pre_apply"`
	PostApply types.String `tfsdk:"post_apply"`
}

func workspaceResourceModelFromAPI(
	ctx context.Context,
	ws *schemas.Workspace,
	pcfgLinks []*schemas.ProviderConfigurationLink,
	stateConsumers []string,
	existing *workspaceResourceModel,
) (*workspaceResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &workspaceResourceModel{
		Id:                        types.StringValue(ws.ID),
		AgentPoolID:               types.StringNull(),
		AutoApply:                 types.BoolValue(ws.Attributes.AutoApply),
		AutoQueueRuns:             types.StringValue(ws.Attributes.AutoQueueRuns),
		CreatedBy:                 types.ListNull(userElementType),
		DeletionProtectionEnabled: types.BoolValue(ws.Attributes.DeletionProtectionEnabled),
		EnvironmentID:             types.StringValue(ws.Relationships.Environment.ID),
		ExecutionMode:             types.StringValue(ws.Attributes.ExecutionMode),
		ForceLatestRun:            types.BoolValue(ws.Attributes.ForceLatestRun),
		HasResources:              types.BoolValue(ws.Attributes.HasResources),
		Hooks:                     types.ListNull(hooksElementType),
		IaCPlatform:               types.StringValue(ws.Attributes.IacPlatform),
		ModuleVersionID:           types.StringNull(),
		Name:                      types.StringValue(ws.Attributes.Name),
		Operations:                types.BoolValue(ws.Attributes.Operations),
		ProviderConfiguration:     types.SetNull(providerConfigurationElementType),
		RemoteBackend:             types.BoolValue(ws.Attributes.RemoteBackend),
		RemoteStateConsumers:      types.SetNull(types.StringType),
		RunOperationTimeout:       types.Int32Null(),
		SSHKeyID:                  types.StringNull(),
		TagIDs:                    types.SetNull(types.StringType),
		TerraformVersion:          types.StringValue(ws.Attributes.TerraformVersion),
		Terragrunt:                types.ListNull(terragruntElementType),
		Type:                      types.StringValue(ws.Attributes.EnvironmentType),
		VarFiles:                  types.ListNull(types.StringType),
		VCSProviderID:             types.StringNull(),
		VCSRepo:                   types.ListNull(vcsRepoElementType),
		WorkingDirectory:          types.StringPointerValue(ws.Attributes.WorkingDirectory),
	}

	if ws.Attributes.VarFiles != nil {
		varFiles, d := types.ListValueFrom(ctx, types.StringType, *ws.Attributes.VarFiles)
		diags.Append(d...)
		model.VarFiles = varFiles
	}

	if ws.Attributes.RunOperationTimeout != nil {
		model.RunOperationTimeout = types.Int32Value(int32(*ws.Attributes.RunOperationTimeout))
	}

	if ws.Relationships.VcsProvider != nil {
		model.VCSProviderID = types.StringValue(ws.Relationships.VcsProvider.ID)
	}

	if ws.Relationships.ModuleVersion != nil {
		model.ModuleVersionID = types.StringValue(ws.Relationships.ModuleVersion.ID)
	}

	if ws.Relationships.SshKey != nil {
		model.SSHKeyID = types.StringValue(ws.Relationships.SshKey.ID)
	}

	if ws.Relationships.AgentPool != nil {
		model.AgentPoolID = types.StringValue(ws.Relationships.AgentPool.ID)
	}

	if ws.Attributes.VcsRepo != nil {
		repo := vcsRepoModel{
			Identifier:        types.StringValue(ws.Attributes.VcsRepo.Identifier),
			Branch:            types.StringPointerValue(ws.Attributes.VcsRepo.Branch),
			Path:              types.StringPointerValue(ws.Attributes.VcsRepo.Path),
			TriggerPatterns:   types.StringPointerValue(ws.Attributes.VcsRepo.TriggerPatterns),
			DryRunsEnabled:    types.BoolValue(ws.Attributes.VcsRepo.DryRunsEnabled),
			IngressSubmodules: types.BoolValue(ws.Attributes.VcsRepo.IngressSubmodules),
			VersionConstraint: types.StringPointerValue(ws.Attributes.VcsRepo.VersionConstraint),
			TriggerPrefixes:   types.ListNull(types.StringType),
		}

		if ws.Attributes.VcsRepo.TriggerPrefixes != nil {
			prefixes, d := types.ListValueFrom(ctx, types.StringType, ws.Attributes.VcsRepo.TriggerPrefixes)
			diags.Append(d...)
			repo.TriggerPrefixes = prefixes
		}

		repoValue, d := types.ListValueFrom(ctx, vcsRepoElementType, []vcsRepoModel{repo})
		diags.Append(d...)
		model.VCSRepo = repoValue
	}

	if ws.Attributes.Terragrunt != nil {
		terragrunt := terragruntModel{
			Version:                     types.StringValue(ws.Attributes.Terragrunt.Version),
			UseRunAll:                   types.BoolValue(ws.Attributes.Terragrunt.UseRunAll),
			IncludeExternalDependencies: types.BoolValue(ws.Attributes.Terragrunt.IncludeExternalDependencies),
		}
		terragruntValue, d := types.ListValueFrom(ctx, terragruntElementType, []terragruntModel{terragrunt})
		diags.Append(d...)
		model.Terragrunt = terragruntValue
	}

	if ws.Relationships.CreatedBy != nil {
		createdBy := []userModel{*userModelFromAPIv2(ws.Relationships.CreatedBy)}
		createdByValue, d := types.ListValueFrom(ctx, userElementType, createdBy)
		diags.Append(d...)
		model.CreatedBy = createdByValue
	}

	var hooks []hooksModel
	if ws.Attributes.Hooks != nil {
		hooks = []hooksModel{{
			PreInit:   types.StringPointerValue(ws.Attributes.Hooks.PreInit),
			PrePlan:   types.StringPointerValue(ws.Attributes.Hooks.PrePlan),
			PostPlan:  types.StringPointerValue(ws.Attributes.Hooks.PostPlan),
			PreApply:  types.StringPointerValue(ws.Attributes.Hooks.PreApply),
			PostApply: types.StringPointerValue(ws.Attributes.Hooks.PostApply),
		}}
	} else if existing != nil && !existing.Hooks.IsNull() {
		hooks = []hooksModel{{
			PreInit:   types.StringValue(""),
			PrePlan:   types.StringValue(""),
			PostPlan:  types.StringValue(""),
			PreApply:  types.StringValue(""),
			PostApply: types.StringValue(""),
		}}
	}
	if len(hooks) > 0 {
		hooksValue, d := types.ListValueFrom(ctx, hooksElementType, hooks)
		diags.Append(d...)
		model.Hooks = hooksValue
	}

	tags := make([]string, len(ws.Relationships.Tags))
	for i, tag := range ws.Relationships.Tags {
		tags[i] = tag.ID
	}
	tagsValue, d := types.SetValueFrom(ctx, types.StringType, tags)
	diags.Append(d...)
	model.TagIDs = tagsValue

	if len(pcfgLinks) > 0 {
		pcfg := make([]providerConfigurationModel, len(pcfgLinks))
		for i, pcfgLink := range pcfgLinks {
			pcfg[i] = providerConfigurationModel{
				ID: types.StringValue(pcfgLink.Relationships.ProviderConfiguration.ID),
				// Set default value to empty string to keep the same behaviour as before
				// (old client would unmarshal into "" even when it is 'null' in the response)
				Alias: types.StringValue(""),
			}
			if pcfgLink.Attributes.Alias != nil {
				pcfg[i].Alias = types.StringValue(*pcfgLink.Attributes.Alias)
			}
		}
		pcfgValue, d := types.SetValueFrom(ctx, providerConfigurationElementType, pcfg)
		diags.Append(d...)
		model.ProviderConfiguration = pcfgValue
	}

	if ws.Attributes.RemoteStateSharing {
		stateConsumers = []string{"*"}
	} else if stateConsumers == nil {
		stateConsumers = []string{}
	}
	consumersValue, d := types.SetValueFrom(ctx, types.StringType, stateConsumers)
	diags.Append(d...)
	model.RemoteStateConsumers = consumersValue

	return model, diags
}
