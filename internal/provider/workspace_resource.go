package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &workspaceResource{}
	_ resource.ResourceWithConfigure        = &workspaceResource{}
	_ resource.ResourceWithConfigValidators = &workspaceResource{}
	_ resource.ResourceWithModifyPlan       = &workspaceResource{}
	_ resource.ResourceWithImportState      = &workspaceResource{}
	_ resource.ResourceWithUpgradeState     = &workspaceResource{}
)

func newWorkspaceResource() resource.Resource {
	return &workspaceResource{}
}

// workspaceResource defines the resource implementation.
type workspaceResource struct {
	framework.ResourceWithScalrClient
}

func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = *workspaceResourceSchema(ctx)
}

func (r *workspaceResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("module_version_id"),
			path.MatchRoot("vcs_provider_id"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("module_version_id"),
			path.MatchRoot("vcs_repo"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("vcs_provider_id"),
			path.MatchRoot("vcs_repo"),
		),
	}
}

func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workspaceResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var varFiles []string
	resp.Diagnostics.Append(plan.VarFiles.ElementsAs(ctx, &varFiles, false)...)

	opts := scalr.WorkspaceCreateOptions{
		AutoApply:                 plan.AutoApply.ValueBoolPointer(),
		AutoQueueRuns:             ptr(scalr.WorkspaceAutoQueueRuns(plan.AutoQueueRuns.ValueString())),
		DeletionProtectionEnabled: plan.DeletionProtectionEnabled.ValueBoolPointer(),
		EnvironmentType:           ptr(scalr.WorkspaceEnvironmentType(plan.Type.ValueString())),
		ExecutionMode:             ptr(scalr.WorkspaceExecutionMode(plan.ExecutionMode.ValueString())),
		ForceLatestRun:            plan.ForceLatestRun.ValueBoolPointer(),
		IacPlatform:               ptr(scalr.WorkspaceIaCPlatform(plan.IaCPlatform.ValueString())),
		Name:                      plan.Name.ValueStringPointer(),
		Operations:                plan.Operations.ValueBoolPointer(),
		VarFiles:                  varFiles,
		WorkingDirectory:          plan.WorkingDirectory.ValueStringPointer(),
		Environment: &scalr.Environment{
			ID: plan.EnvironmentID.ValueString(),
		},
	}

	if !plan.TerraformVersion.IsUnknown() && !plan.TerraformVersion.IsNull() {
		opts.TerraformVersion = plan.TerraformVersion.ValueStringPointer()
	}

	if !plan.RunOperationTimeout.IsUnknown() && !plan.RunOperationTimeout.IsNull() {
		opts.RunOperationTimeout = ptr(int(plan.RunOperationTimeout.ValueInt32()))
	}

	if !plan.VCSProviderID.IsUnknown() && !plan.VCSProviderID.IsNull() {
		opts.VcsProvider = &scalr.VcsProvider{
			ID: plan.VCSProviderID.ValueString(),
		}
	}

	if !plan.ModuleVersionID.IsUnknown() && !plan.ModuleVersionID.IsNull() {
		opts.ModuleVersion = &scalr.ModuleVersion{
			ID: plan.ModuleVersionID.ValueString(),
		}
	}

	if !plan.AgentPoolID.IsUnknown() && !plan.AgentPoolID.IsNull() {
		opts.AgentPool = &scalr.AgentPool{
			ID: plan.AgentPoolID.ValueString(),
		}
	}

	if !plan.VCSRepo.IsUnknown() && !plan.VCSRepo.IsNull() {
		var vcsRepo []vcsRepoModel
		resp.Diagnostics.Append(plan.VCSRepo.ElementsAs(ctx, &vcsRepo, false)...)

		if len(vcsRepo) > 0 {
			repo := vcsRepo[0]

			opts.VCSRepo = &scalr.WorkspaceVCSRepoOptions{
				Identifier:        repo.Identifier.ValueStringPointer(),
				Path:              repo.Path.ValueStringPointer(),
				TriggerPatterns:   repo.TriggerPatterns.ValueStringPointer(),
				DryRunsEnabled:    repo.DryRunsEnabled.ValueBoolPointer(),
				IngressSubmodules: repo.IngressSubmodules.ValueBoolPointer(),
			}

			if !repo.Branch.IsUnknown() && !repo.Branch.IsNull() {
				opts.VCSRepo.Branch = repo.Branch.ValueStringPointer()
			}

			if !repo.TriggerPrefixes.IsUnknown() && !repo.TriggerPrefixes.IsNull() {
				var prefixes []string
				resp.Diagnostics.Append(repo.TriggerPrefixes.ElementsAs(ctx, &prefixes, false)...)
				opts.VCSRepo.TriggerPrefixes = &prefixes
			}
		}
	}

	if !plan.Terragrunt.IsUnknown() && !plan.Terragrunt.IsNull() {
		var terragrunt []terragruntModel
		resp.Diagnostics.Append(plan.Terragrunt.ElementsAs(ctx, &terragrunt, false)...)

		if len(terragrunt) > 0 {
			terr := terragrunt[0]
			opts.Terragrunt = &scalr.WorkspaceTerragruntOptions{
				Version:                     terr.Version.ValueString(),
				UseRunAll:                   terr.UseRunAll.ValueBoolPointer(),
				IncludeExternalDependencies: terr.IncludeExternalDependencies.ValueBoolPointer(),
			}
		}
	}

	if !plan.Hooks.IsUnknown() && !plan.Hooks.IsNull() {
		var hooks []hooksModel
		resp.Diagnostics.Append(plan.Hooks.ElementsAs(ctx, &hooks, false)...)

		if len(hooks) > 0 {
			hook := hooks[0]
			opts.Hooks = &scalr.HooksOptions{
				PreInit:   hook.PreInit.ValueStringPointer(),
				PrePlan:   hook.PrePlan.ValueStringPointer(),
				PostPlan:  hook.PostPlan.ValueStringPointer(),
				PreApply:  hook.PreApply.ValueStringPointer(),
				PostApply: hook.PostApply.ValueStringPointer(),
			}
		}
	}

	if !plan.TagIDs.IsUnknown() && !plan.TagIDs.IsNull() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagIDs.ElementsAs(ctx, &tagIDs, false)...)

		tags := make([]*scalr.Tag, len(tagIDs))
		for i, tagID := range tagIDs {
			tags[i] = &scalr.Tag{ID: tagID}
		}

		opts.Tags = tags
	}

	remoteStateConsumers := make([]*scalr.WorkspaceRelation, 0)
	if !plan.RemoteStateConsumers.IsUnknown() && !plan.RemoteStateConsumers.IsNull() {
		opts.RemoteStateSharing = ptr(false)
		var consumerIDs []string
		resp.Diagnostics.Append(plan.RemoteStateConsumers.ElementsAs(ctx, &consumerIDs, false)...)

		if (len(consumerIDs) == 1) && (consumerIDs[0] == "*") {
			opts.RemoteStateSharing = ptr(true)
		} else if len(consumerIDs) > 0 {
			for _, consumerID := range consumerIDs {
				remoteStateConsumers = append(remoteStateConsumers, &scalr.WorkspaceRelation{ID: consumerID})
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	workspace, err := r.Client.Workspaces.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workspace", err.Error())
		return
	}

	if !plan.ProviderConfiguration.IsUnknown() && !plan.ProviderConfiguration.IsNull() {
		var pcfgs []providerConfigurationModel
		resp.Diagnostics.Append(plan.ProviderConfiguration.ElementsAs(ctx, &pcfgs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, pcfg := range pcfgs {
			pcfgOpts := scalr.ProviderConfigurationLinkCreateOptions{
				ProviderConfiguration: &scalr.ProviderConfiguration{ID: pcfg.ID.ValueString()},
			}
			if !pcfg.Alias.IsUnknown() && len(pcfg.Alias.ValueString()) > 0 {
				pcfgOpts.Alias = pcfg.Alias.ValueStringPointer()
			}
			_, err = r.Client.ProviderConfigurationLinks.Create(ctx, workspace.ID, pcfgOpts)
			if err != nil {
				resp.Diagnostics.AddError("Error creating provider configuration link", err.Error())
				return
			}
		}
	}

	if !plan.SSHKeyID.IsUnknown() && !plan.SSHKeyID.IsNull() {
		_, err = r.Client.SSHKeysLinks.Create(ctx, workspace.ID, plan.SSHKeyID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error creating SSH key link", err.Error())
			return
		}
	}

	if len(remoteStateConsumers) > 0 {
		err = r.Client.RemoteStateConsumers.Add(ctx, workspace.ID, remoteStateConsumers)
		if err != nil {
			resp.Diagnostics.AddError("Error adding remote state consumers", err.Error())
			return
		}
	}

	// Get refreshed resource state from API
	workspace, err = r.Client.Workspaces.ReadByID(ctx, workspace.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.Client, workspace.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
	}

	stateConsumers, err := getRemoteStateConsumers(ctx, r.Client, workspace.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
	}

	result, diags := workspaceResourceModelFromAPI(ctx, workspace, pcfgLinks, stateConsumers, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspace, err := r.Client.Workspaces.ReadByID(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.Client, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
	}

	stateConsumers, err := getRemoteStateConsumers(ctx, r.Client, workspace.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
	}

	result, diags := workspaceResourceModelFromAPI(ctx, workspace, pcfgLinks, stateConsumers, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state workspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.WorkspaceUpdateOptions{}

	if !plan.Name.Equal(state.Name) {
		opts.Name = plan.Name.ValueStringPointer()
	}

	if !plan.AutoApply.Equal(state.AutoApply) {
		opts.AutoApply = plan.AutoApply.ValueBoolPointer()
	}

	if !plan.AutoQueueRuns.Equal(state.AutoQueueRuns) {
		opts.AutoQueueRuns = ptr(scalr.WorkspaceAutoQueueRuns(plan.AutoQueueRuns.ValueString()))
	}

	if !plan.DeletionProtectionEnabled.Equal(state.DeletionProtectionEnabled) {
		opts.DeletionProtectionEnabled = plan.DeletionProtectionEnabled.ValueBoolPointer()
	}

	if !plan.ExecutionMode.Equal(state.ExecutionMode) {
		opts.ExecutionMode = ptr(scalr.WorkspaceExecutionMode(plan.ExecutionMode.ValueString()))
	}

	if !plan.ForceLatestRun.Equal(state.ForceLatestRun) {
		opts.ForceLatestRun = plan.ForceLatestRun.ValueBoolPointer()
	}

	if !plan.IaCPlatform.Equal(state.IaCPlatform) {
		opts.IacPlatform = ptr(scalr.WorkspaceIaCPlatform(plan.IaCPlatform.ValueString()))
	}

	if !plan.Operations.Equal(state.Operations) {
		opts.Operations = plan.Operations.ValueBoolPointer()
	}

	if !plan.RunOperationTimeout.Equal(state.RunOperationTimeout) && !plan.RunOperationTimeout.IsNull() {
		opts.RunOperationTimeout = ptr(int(plan.RunOperationTimeout.ValueInt32()))
	}

	if !plan.TerraformVersion.Equal(state.TerraformVersion) && !plan.TerraformVersion.IsNull() {
		opts.TerraformVersion = plan.TerraformVersion.ValueStringPointer()
	}

	if !plan.Type.Equal(state.Type) {
		opts.EnvironmentType = ptr(scalr.WorkspaceEnvironmentType(plan.Type.ValueString()))
	}

	if !plan.WorkingDirectory.Equal(state.WorkingDirectory) {
		opts.WorkingDirectory = plan.WorkingDirectory.ValueStringPointer()
	}

	if !plan.VCSProviderID.IsNull() {
		opts.VcsProvider = &scalr.VcsProvider{ID: plan.VCSProviderID.ValueString()}
	}

	if !plan.ModuleVersionID.IsNull() {
		opts.ModuleVersion = &scalr.ModuleVersion{ID: plan.ModuleVersionID.ValueString()}
	}

	if !plan.AgentPoolID.IsNull() {
		opts.AgentPool = &scalr.AgentPool{ID: plan.AgentPoolID.ValueString()}
	}

	if !plan.VCSRepo.IsNull() {
		var vcsRepo []vcsRepoModel
		resp.Diagnostics.Append(plan.VCSRepo.ElementsAs(ctx, &vcsRepo, false)...)

		if len(vcsRepo) > 0 {
			repo := vcsRepo[0]

			opts.VCSRepo = &scalr.WorkspaceVCSRepoOptions{
				Identifier:        repo.Identifier.ValueStringPointer(),
				Path:              repo.Path.ValueStringPointer(),
				TriggerPatterns:   repo.TriggerPatterns.ValueStringPointer(),
				DryRunsEnabled:    repo.DryRunsEnabled.ValueBoolPointer(),
				IngressSubmodules: repo.IngressSubmodules.ValueBoolPointer(),
			}

			if !repo.Branch.IsUnknown() && !repo.Branch.IsNull() {
				opts.VCSRepo.Branch = repo.Branch.ValueStringPointer()
			}

			if !repo.TriggerPrefixes.IsUnknown() && !repo.TriggerPrefixes.IsNull() {
				var prefixes []string
				resp.Diagnostics.Append(repo.TriggerPrefixes.ElementsAs(ctx, &prefixes, false)...)
				opts.VCSRepo.TriggerPrefixes = &prefixes
			}
		}
	}

	if !plan.Terragrunt.IsNull() {
		var terragrunt []terragruntModel
		resp.Diagnostics.Append(plan.Terragrunt.ElementsAs(ctx, &terragrunt, false)...)

		if len(terragrunt) > 0 {
			terr := terragrunt[0]
			opts.Terragrunt = &scalr.WorkspaceTerragruntOptions{
				Version:                     terr.Version.ValueString(),
				UseRunAll:                   terr.UseRunAll.ValueBoolPointer(),
				IncludeExternalDependencies: terr.IncludeExternalDependencies.ValueBoolPointer(),
			}
		}
	}

	if !plan.Hooks.Equal(state.Hooks) {
		var hooks []hooksModel
		resp.Diagnostics.Append(plan.Hooks.ElementsAs(ctx, &hooks, false)...)

		if len(hooks) > 0 {
			hook := hooks[0]
			opts.Hooks = &scalr.HooksOptions{
				PreInit:   hook.PreInit.ValueStringPointer(),
				PrePlan:   hook.PrePlan.ValueStringPointer(),
				PostPlan:  hook.PostPlan.ValueStringPointer(),
				PreApply:  hook.PreApply.ValueStringPointer(),
				PostApply: hook.PostApply.ValueStringPointer(),
			}
		} else {
			opts.Hooks = &scalr.HooksOptions{
				PreInit:   ptr(""),
				PrePlan:   ptr(""),
				PostPlan:  ptr(""),
				PreApply:  ptr(""),
				PostApply: ptr(""),
			}
		}
	}

	if !plan.VarFiles.IsNull() {
		var varFiles []string
		resp.Diagnostics.Append(plan.VarFiles.ElementsAs(ctx, &varFiles, false)...)
		opts.VarFiles = varFiles
	}

	var consumersToAdd, consumersToRemove []string
	if !plan.RemoteStateConsumers.Equal(state.RemoteStateConsumers) {
		var planConsumers []string
		var stateConsumers []string
		resp.Diagnostics.Append(plan.RemoteStateConsumers.ElementsAs(ctx, &planConsumers, false)...)
		resp.Diagnostics.Append(state.RemoteStateConsumers.ElementsAs(ctx, &stateConsumers, false)...)

		opts.RemoteStateSharing = ptr(false)

		if len(planConsumers) == 1 && planConsumers[0] == "*" {
			opts.RemoteStateSharing = ptr(true)
			planConsumers = []string{}
		}
		if len(stateConsumers) == 1 && stateConsumers[0] == "*" {
			stateConsumers = []string{}
		}

		consumersToAdd, consumersToRemove = diff(stateConsumers, planConsumers)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing resource
	_, err := r.Client.Workspaces.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating workspace", err.Error())
		return
	}

	if !plan.TagIDs.Equal(state.TagIDs) {
		var planTags []string
		var stateTags []string
		resp.Diagnostics.Append(plan.TagIDs.ElementsAs(ctx, &planTags, false)...)
		resp.Diagnostics.Append(state.TagIDs.ElementsAs(ctx, &stateTags, false)...)

		tagsToAdd, tagsToRemove := diff(stateTags, planTags)

		if len(tagsToAdd) > 0 {
			tagRelations := make([]*scalr.TagRelation, len(tagsToAdd))
			for i, tag := range tagsToAdd {
				tagRelations[i] = &scalr.TagRelation{ID: tag}
			}
			err = r.Client.WorkspaceTags.Add(ctx, plan.Id.ValueString(), tagRelations)
			if err != nil {
				resp.Diagnostics.AddError("Error adding tags to workspace", err.Error())
			}
		}

		if len(tagsToRemove) > 0 {
			tagRelations := make([]*scalr.TagRelation, len(tagsToRemove))
			for i, tag := range tagsToRemove {
				tagRelations[i] = &scalr.TagRelation{ID: tag}
			}
			err = r.Client.WorkspaceTags.Delete(ctx, plan.Id.ValueString(), tagRelations)
			if err != nil {
				resp.Diagnostics.AddError("Error removing tags from workspace", err.Error())
			}
		}
	}

	if !plan.ProviderConfiguration.Equal(state.ProviderConfiguration) {
		expectedLinks := make(map[string]scalr.ProviderConfigurationLinkCreateOptions)
		if !plan.ProviderConfiguration.IsNull() {
			var pcfgs []providerConfigurationModel
			resp.Diagnostics.Append(plan.ProviderConfiguration.ElementsAs(ctx, &pcfgs, false)...)

			for _, pcfg := range pcfgs {
				mapID := pcfg.ID.ValueString()
				pcfgOpts := scalr.ProviderConfigurationLinkCreateOptions{
					ProviderConfiguration: &scalr.ProviderConfiguration{ID: pcfg.ID.ValueString()},
				}
				if !pcfg.Alias.IsUnknown() && len(pcfg.Alias.ValueString()) > 0 {
					pcfgOpts.Alias = pcfg.Alias.ValueStringPointer()
					mapID = mapID + pcfg.Alias.ValueString()
				}
				expectedLinks[mapID] = pcfgOpts
			}
		}

		currentLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.Client, plan.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
		} else {
			for _, currentLink := range currentLinks {
				mapID := currentLink.ProviderConfiguration.ID + currentLink.Alias
				if _, ok := expectedLinks[mapID]; ok {
					delete(expectedLinks, mapID)
				} else {
					err = r.Client.ProviderConfigurationLinks.Delete(ctx, currentLink.ID)
					if err != nil {
						resp.Diagnostics.AddError("Error deleting provider configuration link", err.Error())
					}
				}
			}

			for _, link := range expectedLinks {
				_, err = r.Client.ProviderConfigurationLinks.Create(ctx, plan.Id.ValueString(), link)
				if err != nil {
					resp.Diagnostics.AddError("Error creating provider configuration link", err.Error())
				}
			}
		}
	}

	if !plan.SSHKeyID.Equal(state.SSHKeyID) {
		if !state.SSHKeyID.IsNull() && plan.SSHKeyID.IsNull() {
			err = r.Client.SSHKeysLinks.Delete(ctx, plan.Id.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error deleting SSH key link", err.Error())
			}
		} else if !plan.SSHKeyID.IsNull() {
			_, err = r.Client.SSHKeysLinks.Create(ctx, plan.Id.ValueString(), plan.SSHKeyID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error creating SSH key link", err.Error())
			}
		}
	}

	if len(consumersToAdd) > 0 {
		c := make([]*scalr.WorkspaceRelation, len(consumersToAdd))
		for i, consumer := range consumersToAdd {
			c[i] = &scalr.WorkspaceRelation{ID: consumer}
		}
		err = r.Client.RemoteStateConsumers.Add(ctx, plan.Id.ValueString(), c)
		if err != nil {
			resp.Diagnostics.AddError("Error adding remote state consumers", err.Error())
		}
	}
	if len(consumersToRemove) > 0 {
		c := make([]*scalr.WorkspaceRelation, len(consumersToRemove))
		for i, consumer := range consumersToRemove {
			c[i] = &scalr.WorkspaceRelation{ID: consumer}
		}
		err = r.Client.RemoteStateConsumers.Delete(ctx, plan.Id.ValueString(), c)
		if err != nil {
			resp.Diagnostics.AddError("Error removing remote state consumers", err.Error())
		}
	}

	// Get refreshed resource state from API
	workspace, err := r.Client.Workspaces.ReadByID(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.Client, workspace.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
	}

	stateConsumers, err := getRemoteStateConsumers(ctx, r.Client, workspace.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
	}

	result, diags := workspaceResourceModelFromAPI(ctx, workspace, pcfgLinks, stateConsumers, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state workspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.Workspaces.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting workspace", err.Error())
		return
	}
}

func (r *workspaceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// The resource is being destroyed
		return
	}

	var operations, operationsCfg types.Bool
	var executionMode, executionModeCfg types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("operations"), &operations)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("operations"), &operationsCfg)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("execution_mode"), &executionMode)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("execution_mode"), &executionModeCfg)...)

	if !operationsCfg.IsNull() && !executionModeCfg.IsNull() {
		// Both attributes cannot be set to mutually exclusive values at the same time.
		if !operations.ValueBool() && executionMode.ValueString() == string(scalr.WorkspaceExecutionModeRemote) ||
			operations.ValueBool() && executionMode.ValueString() == string(scalr.WorkspaceExecutionModeLocal) {
			resp.Diagnostics.AddError(
				"Attributes `operations` and `execution_mode` are configured with conflicting values",
				"The attribute `operations` is deprecated. Use `execution_mode` instead",
			)
		}
	}

	if !operationsCfg.IsNull() && executionModeCfg.IsNull() {
		// When the `operations` is explicitly set in the configuration, and `execution_mode` is not -
		// this is the only case when it takes precedence.
		if !operations.ValueBool() {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("execution_mode"), string(scalr.WorkspaceExecutionModeLocal))...)
		} else {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("execution_mode"), string(scalr.WorkspaceExecutionModeRemote))...)
		}
	} else {
		// In all other cases, `execution_mode` dictates the value for `operations`, even when left default.
		if executionMode.ValueString() == string(scalr.WorkspaceExecutionModeLocal) {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("operations"), false)...)
		} else {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("operations"), true)...)
		}
	}
}

func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *workspaceResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   workspaceResourceSchemaV0(ctx),
			StateUpgrader: upgradeWorkspaceResourceStateV0toV4(r.Client),
		},
		1: {
			PriorSchema:   workspaceResourceSchemaV1(ctx),
			StateUpgrader: upgradeWorkspaceResourceStateV1toV4(r.Client),
		},
		2: {
			PriorSchema:   workspaceResourceSchemaV2(ctx),
			StateUpgrader: upgradeWorkspaceResourceStateV2toV4(r.Client),
		},
		3: {
			PriorSchema:   workspaceResourceSchemaV3(ctx),
			StateUpgrader: upgradeWorkspaceResourceStateV3toV4(r.Client),
		},
	}
}

func getProviderConfigurationWorkspaceLinks(
	ctx context.Context, scalrClient *scalr.Client, workspaceId string,
) (workspaceLinks []*scalr.ProviderConfigurationLink, err error) {
	linkListOption := scalr.ProviderConfigurationLinksListOptions{Include: "provider-configuration"}
	for {
		linksList, err := scalrClient.ProviderConfigurationLinks.List(ctx, workspaceId, linkListOption)

		if err != nil {
			return nil, err
		}

		for _, link := range linksList.Items {
			if link.Workspace != nil {
				workspaceLinks = append(workspaceLinks, link)
			}
		}

		// Exit the loop when we've seen all pages.
		if linksList.CurrentPage >= linksList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		linkListOption.PageNumber = linksList.NextPage
	}
	return
}

func getRemoteStateConsumers(
	ctx context.Context, scalrClient *scalr.Client, workspaceId string,
) (consumers []string, err error) {
	listOpts := scalr.ListOptions{}
	for {
		cl, err := scalrClient.RemoteStateConsumers.List(ctx, workspaceId, listOpts)
		if err != nil {
			return nil, err
		}

		for _, c := range cl.Items {
			consumers = append(consumers, c.ID)
		}

		if cl.CurrentPage >= cl.TotalPages {
			break
		}
		listOpts.PageNumber = cl.NextPage
	}
	return
}
