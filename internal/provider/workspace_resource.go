package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/ops/environment"
	"github.com/scalr/go-scalr/v2/scalr/ops/workspace"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

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

	opts := schemas.WorkspaceRequest{
		Attributes: schemas.WorkspaceAttributesRequest{
			AutoApply:                 value.SetPtrMaybe(plan.AutoApply.ValueBoolPointer()),
			DeletionProtectionEnabled: value.SetPtrMaybe(plan.DeletionProtectionEnabled.ValueBoolPointer()),
			EnvironmentType:           value.Set(schemas.WorkspaceEnvironmentType(plan.Type.ValueString())),
			ExecutionMode:             value.Set(schemas.WorkspaceExecutionMode(plan.ExecutionMode.ValueString())),
			ForceLatestRun:            value.SetPtrMaybe(plan.ForceLatestRun.ValueBoolPointer()),
			IacPlatform:               value.Set(schemas.WorkspaceIacPlatform(plan.IaCPlatform.ValueString())),
			Name:                      value.Set(plan.Name.ValueString()),
			Operations:                value.SetPtrMaybe(plan.Operations.ValueBoolPointer()),
			VarFiles:                  value.Set(varFiles),
			WorkingDirectory:          value.SetPtrMaybe(plan.WorkingDirectory.ValueStringPointer()),
		},
		Relationships: schemas.WorkspaceRelationshipsRequest{
			Environment: value.Set(
				schemas.Environment{
					ID: plan.EnvironmentID.ValueString(),
				},
			),
		},
	}

	if !plan.AutoQueueRuns.IsUnknown() && !plan.AutoQueueRuns.IsNull() {
		opts.Attributes.AutoQueueRuns = value.Set(schemas.WorkspaceAutoQueueRuns(plan.AutoQueueRuns.ValueString()))
	}

	if !plan.TerraformVersion.IsUnknown() && !plan.TerraformVersion.IsNull() {
		opts.Attributes.TerraformVersion = value.Set(plan.TerraformVersion.ValueString())
	}

	if !plan.RunOperationTimeout.IsUnknown() && !plan.RunOperationTimeout.IsNull() {
		opts.Attributes.RunOperationTimeout = value.Set(int(plan.RunOperationTimeout.ValueInt32()))
	}

	if !plan.RemoteBackend.IsUnknown() && !plan.RemoteBackend.IsNull() {
		opts.Attributes.RemoteBackend = value.Set(plan.RemoteBackend.ValueBool())
	}

	if !plan.VCSProviderID.IsUnknown() && !plan.VCSProviderID.IsNull() {
		opts.Relationships.VcsProvider = value.Set(
			schemas.VcsProvider{
				ID: plan.VCSProviderID.ValueString(),
			},
		)
	}

	if !plan.ModuleVersionID.IsUnknown() && !plan.ModuleVersionID.IsNull() {
		opts.Relationships.ModuleVersion = value.Set(
			schemas.ModuleVersion{
				ID: plan.ModuleVersionID.ValueString(),
			},
		)
	}

	if !plan.AgentPoolID.IsUnknown() && !plan.AgentPoolID.IsNull() {
		opts.Relationships.AgentPool = value.Set(
			schemas.AgentPool{
				ID: plan.AgentPoolID.ValueString(),
			},
		)
	}

	if !plan.VCSRepo.IsUnknown() && !plan.VCSRepo.IsNull() {
		var vcsRepo []vcsRepoModel
		resp.Diagnostics.Append(plan.VCSRepo.ElementsAs(ctx, &vcsRepo, false)...)

		if len(vcsRepo) > 0 {
			repo := vcsRepo[0]

			vcsRepoOpts := schemas.WorkspaceVcsRepoRequest{
				Identifier:        value.Set(repo.Identifier.ValueString()),
				Path:              value.Set(repo.Path.ValueString()),
				TriggerPatterns:   value.Set(repo.TriggerPatterns.ValueString()),
				DryRunsEnabled:    value.Set(repo.DryRunsEnabled.ValueBool()),
				IngressSubmodules: value.Set(repo.IngressSubmodules.ValueBool()),
			}

			if !repo.Branch.IsUnknown() && !repo.Branch.IsNull() {
				vcsRepoOpts.Branch = value.Set(repo.Branch.ValueString())
			}

			if !repo.VersionConstraint.IsUnknown() && !repo.VersionConstraint.IsNull() {
				vcsRepoOpts.VersionConstraint = value.Set(repo.VersionConstraint.ValueString())
			}

			if !repo.TriggerPrefixes.IsUnknown() && !repo.TriggerPrefixes.IsNull() {
				var prefixes []string
				resp.Diagnostics.Append(repo.TriggerPrefixes.ElementsAs(ctx, &prefixes, false)...)
				vcsRepoOpts.TriggerPrefixes = value.Set(prefixes)
			}

			opts.Attributes.VcsRepo = value.Set(vcsRepoOpts)
		}
	}

	if !plan.Terragrunt.IsUnknown() && !plan.Terragrunt.IsNull() {
		var terragrunt []terragruntModel
		resp.Diagnostics.Append(plan.Terragrunt.ElementsAs(ctx, &terragrunt, false)...)

		if len(terragrunt) > 0 {
			terr := terragrunt[0]
			opts.Attributes.Terragrunt = value.Set(
				schemas.WorkspaceTerragruntRequest{
					Version:                     value.Set(terr.Version.ValueString()),
					UseRunAll:                   value.SetPtrMaybe(terr.UseRunAll.ValueBoolPointer()),
					IncludeExternalDependencies: value.SetPtrMaybe(terr.IncludeExternalDependencies.ValueBoolPointer()),
				},
			)
		}
	}

	if !plan.Hooks.IsUnknown() && !plan.Hooks.IsNull() {
		var hooks []hooksModel
		resp.Diagnostics.Append(plan.Hooks.ElementsAs(ctx, &hooks, false)...)

		if len(hooks) > 0 {
			hook := hooks[0]
			opts.Attributes.Hooks = value.Set(
				schemas.WorkspaceHooksRequest{
					PreInit:   value.SetPtrMaybe(hook.PreInit.ValueStringPointer()),
					PrePlan:   value.SetPtrMaybe(hook.PrePlan.ValueStringPointer()),
					PostPlan:  value.SetPtrMaybe(hook.PostPlan.ValueStringPointer()),
					PreApply:  value.SetPtrMaybe(hook.PreApply.ValueStringPointer()),
					PostApply: value.SetPtrMaybe(hook.PostApply.ValueStringPointer()),
				},
			)
		}
	}

	if !plan.TagIDs.IsUnknown() && !plan.TagIDs.IsNull() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagIDs.ElementsAs(ctx, &tagIDs, false)...)

		tags := make([]schemas.Tag, len(tagIDs))
		for i, tagID := range tagIDs {
			tags[i] = schemas.Tag{ID: tagID}
		}

		opts.Relationships.Tags = value.Set(tags)
	}

	remoteStateConsumers := make([]schemas.Workspace, 0)
	if !plan.RemoteStateConsumers.IsUnknown() && !plan.RemoteStateConsumers.IsNull() {
		opts.Attributes.RemoteStateSharing = value.Set(false)
		var consumerIDs []string
		resp.Diagnostics.Append(plan.RemoteStateConsumers.ElementsAs(ctx, &consumerIDs, false)...)

		if (len(consumerIDs) == 1) && (consumerIDs[0] == "*") {
			opts.Attributes.RemoteStateSharing = value.Set(true)
		} else if len(consumerIDs) > 0 {
			for _, consumerID := range consumerIDs {
				remoteStateConsumers = append(remoteStateConsumers, schemas.Workspace{ID: consumerID})
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ws, err := r.ClientV2.Workspace.CreateWorkspace(ctx, &opts)
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
			pcfgOpts := schemas.ProviderConfigurationLinkRequest{
				Relationships: schemas.ProviderConfigurationLinkRelationshipsRequest{
					ProviderConfiguration: value.Set(schemas.ProviderConfiguration{ID: pcfg.ID.ValueString()}),
				},
			}
			if !pcfg.Alias.IsUnknown() && len(pcfg.Alias.ValueString()) > 0 {
				pcfgOpts.Attributes.Alias = value.Set(pcfg.Alias.ValueString())
			}
			_, err = r.ClientV2.ProviderConfigurationLink.CreateProviderConfigurationLink(ctx, ws.ID, &pcfgOpts)
			if err != nil {
				resp.Diagnostics.AddError("Error creating provider configuration link", err.Error())
				return
			}
		}
	}

	if !plan.SSHKeyID.IsUnknown() && !plan.SSHKeyID.IsNull() {
		_, err = r.ClientV2.Misc.CreateWorkspaceSshKeyLink(
			ctx,
			ws.ID,
			&schemas.WorkspaceSSHKeyLinkRequest{SshKey: plan.SSHKeyID.ValueString()},
		)
		if err != nil {
			resp.Diagnostics.AddError("Error creating SSH key link", err.Error())
			return
		}
	}

	if len(remoteStateConsumers) > 0 {
		err = r.ClientV2.Workspace.AddRemoteStateConsumers(ctx, ws.ID, remoteStateConsumers)
		if err != nil {
			resp.Diagnostics.AddError("Error adding remote state consumers", err.Error())
			return
		}
	}

	// Get refreshed resource state from API
	ws, err = r.ClientV2.Workspace.GetWorkspace(
		ctx, ws.ID, &workspace.GetWorkspaceOptions{
			Include: []string{"created-by"},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.ClientV2, ws.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
	}

	stateConsumers, err := getRemoteStateConsumers(ctx, r.ClientV2, ws.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
	}

	result, diags := workspaceResourceModelFromAPI(ctx, ws, pcfgLinks, stateConsumers, &plan)
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

	ws, err := r.ClientV2.Workspace.GetWorkspace(
		ctx, state.Id.ValueString(), &workspace.GetWorkspaceOptions{
			Include: []string{"created-by"},
		},
	)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.ClientV2, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
	}

	stateConsumers, err := getRemoteStateConsumers(ctx, r.ClientV2, ws.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
	}

	result, diags := workspaceResourceModelFromAPI(ctx, ws, pcfgLinks, stateConsumers, &state)
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

	opts := schemas.WorkspaceRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.AutoApply.Equal(state.AutoApply) {
		opts.Attributes.AutoApply = value.Set(plan.AutoApply.ValueBool())
	}

	if !plan.AutoQueueRuns.Equal(state.AutoQueueRuns) && !plan.AutoQueueRuns.IsNull() && !plan.AutoQueueRuns.IsUnknown() {
		opts.Attributes.AutoQueueRuns = value.Set(schemas.WorkspaceAutoQueueRuns(plan.AutoQueueRuns.ValueString()))
	}

	if !plan.DeletionProtectionEnabled.Equal(state.DeletionProtectionEnabled) {
		opts.Attributes.DeletionProtectionEnabled = value.Set(plan.DeletionProtectionEnabled.ValueBool())
	}

	if !plan.ExecutionMode.Equal(state.ExecutionMode) {
		opts.Attributes.ExecutionMode = value.Set(schemas.WorkspaceExecutionMode(plan.ExecutionMode.ValueString()))
	}

	if !plan.ForceLatestRun.Equal(state.ForceLatestRun) {
		opts.Attributes.ForceLatestRun = value.Set(plan.ForceLatestRun.ValueBool())
	}

	if !plan.IaCPlatform.Equal(state.IaCPlatform) {
		opts.Attributes.IacPlatform = value.Set(schemas.WorkspaceIacPlatform(plan.IaCPlatform.ValueString()))
	}

	if !plan.Operations.Equal(state.Operations) {
		opts.Attributes.Operations = value.Set(plan.Operations.ValueBool())
	}

	if !plan.RunOperationTimeout.Equal(state.RunOperationTimeout) {
		if plan.RunOperationTimeout.IsNull() {
			// RunOperationTimeout can be explicitly set to null via API
			opts.Attributes.RunOperationTimeout = value.Null[int]()
		} else {
			opts.Attributes.RunOperationTimeout = value.Set(int(plan.RunOperationTimeout.ValueInt32()))
		}
	}

	if !plan.TerraformVersion.Equal(state.TerraformVersion) {
		opts.Attributes.TerraformVersion = framework.SetIfKnownString(plan.TerraformVersion)
	}

	if !plan.Type.Equal(state.Type) {
		opts.Attributes.EnvironmentType = value.Set(schemas.WorkspaceEnvironmentType(plan.Type.ValueString()))
	}

	if !plan.WorkingDirectory.Equal(state.WorkingDirectory) {
		opts.Attributes.WorkingDirectory = value.Set(plan.WorkingDirectory.ValueString())
	}

	if !plan.VCSProviderID.Equal(state.VCSProviderID) {
		if plan.VCSProviderID.IsNull() {
			opts.Relationships.VcsProvider = value.Null[schemas.VcsProvider]()
		} else {
			opts.Relationships.VcsProvider = value.Set(schemas.VcsProvider{ID: plan.VCSProviderID.ValueString()})
		}
	}

	if !plan.ModuleVersionID.Equal(state.ModuleVersionID) {
		if plan.ModuleVersionID.IsNull() {
			opts.Relationships.ModuleVersion = value.Null[schemas.ModuleVersion]()
		} else {
			opts.Relationships.ModuleVersion = value.Set(schemas.ModuleVersion{ID: plan.ModuleVersionID.ValueString()})
		}
	}

	if !plan.AgentPoolID.Equal(state.AgentPoolID) {
		if plan.AgentPoolID.IsNull() {
			opts.Relationships.AgentPool = value.Null[schemas.AgentPool]()
		} else {
			opts.Relationships.AgentPool = value.Set(schemas.AgentPool{ID: plan.AgentPoolID.ValueString()})
		}
	}

	if !plan.VCSRepo.Equal(state.VCSRepo) {
		if plan.VCSRepo.IsNull() {
			opts.Attributes.VcsRepo = value.Null[schemas.WorkspaceVcsRepoRequest]()
		} else {
			var vcsRepo []vcsRepoModel
			resp.Diagnostics.Append(plan.VCSRepo.ElementsAs(ctx, &vcsRepo, false)...)

			if len(vcsRepo) > 0 {
				repo := vcsRepo[0]

				vcsRepoOpts := schemas.WorkspaceVcsRepoRequest{
					Identifier:        value.Set(repo.Identifier.ValueString()),
					Path:              value.Set(repo.Path.ValueString()),
					TriggerPatterns:   value.Set(repo.TriggerPatterns.ValueString()),
					DryRunsEnabled:    value.Set(repo.DryRunsEnabled.ValueBool()),
					IngressSubmodules: value.Set(repo.IngressSubmodules.ValueBool()),
					// Branch and VersionConstraint are optional and mutually exclusive.
					// API requires one to be explicitly set to null if another is set,
					// Pre-nullify the values so neither is left Unset.
					Branch:            value.Null[string](),
					VersionConstraint: value.Null[string](),
				}

				if !repo.Branch.IsUnknown() && !repo.Branch.IsNull() {
					vcsRepoOpts.Branch = value.Set(repo.Branch.ValueString())
				}

				if !repo.VersionConstraint.IsUnknown() && !repo.VersionConstraint.IsNull() {
					vcsRepoOpts.VersionConstraint = value.Set(repo.VersionConstraint.ValueString())
				}

				if !repo.TriggerPrefixes.IsUnknown() && !repo.TriggerPrefixes.IsNull() {
					var prefixes []string
					resp.Diagnostics.Append(repo.TriggerPrefixes.ElementsAs(ctx, &prefixes, false)...)
					vcsRepoOpts.TriggerPrefixes = value.Set(prefixes)
				}

				opts.Attributes.VcsRepo = value.Set(vcsRepoOpts)
			}
		}
	}

	if !plan.Terragrunt.Equal(state.Terragrunt) {
		if plan.Terragrunt.IsNull() {
			opts.Attributes.Terragrunt = value.Null[schemas.WorkspaceTerragruntRequest]()
		} else {
			var terragrunt []terragruntModel
			resp.Diagnostics.Append(plan.Terragrunt.ElementsAs(ctx, &terragrunt, false)...)

			if len(terragrunt) > 0 {
				terr := terragrunt[0]
				opts.Attributes.Terragrunt = value.Set(
					schemas.WorkspaceTerragruntRequest{
						Version:                     value.Set(terr.Version.ValueString()),
						UseRunAll:                   value.SetPtrMaybe(terr.UseRunAll.ValueBoolPointer()),
						IncludeExternalDependencies: value.SetPtrMaybe(terr.IncludeExternalDependencies.ValueBoolPointer()),
					},
				)
			}
		}
	}

	if !plan.Hooks.Equal(state.Hooks) {
		if plan.Hooks.IsNull() {
			opts.Attributes.Hooks = value.Null[schemas.WorkspaceHooksRequest]()
		} else {
			var hooks []hooksModel
			resp.Diagnostics.Append(plan.Hooks.ElementsAs(ctx, &hooks, false)...)

			if len(hooks) > 0 {
				hook := hooks[0]
				opts.Attributes.Hooks = value.Set(
					schemas.WorkspaceHooksRequest{
						PreInit:   value.SetPtrMaybe(hook.PreInit.ValueStringPointer()),
						PrePlan:   value.SetPtrMaybe(hook.PrePlan.ValueStringPointer()),
						PostPlan:  value.SetPtrMaybe(hook.PostPlan.ValueStringPointer()),
						PreApply:  value.SetPtrMaybe(hook.PreApply.ValueStringPointer()),
						PostApply: value.SetPtrMaybe(hook.PostApply.ValueStringPointer()),
					},
				)
			}
		}
	}

	if !plan.VarFiles.Equal(state.VarFiles) {
		if plan.VarFiles.IsNull() {
			opts.Attributes.VarFiles = value.Null[[]string]()
		} else {
			var varFiles []string
			resp.Diagnostics.Append(plan.VarFiles.ElementsAs(ctx, &varFiles, false)...)
			opts.Attributes.VarFiles = value.Set(varFiles)
		}
	}

	var consumersToAdd, consumersToRemove []string
	if !plan.RemoteStateConsumers.Equal(state.RemoteStateConsumers) {
		var planConsumers []string
		var stateConsumers []string
		resp.Diagnostics.Append(plan.RemoteStateConsumers.ElementsAs(ctx, &planConsumers, false)...)
		resp.Diagnostics.Append(state.RemoteStateConsumers.ElementsAs(ctx, &stateConsumers, false)...)

		opts.Attributes.RemoteStateSharing = value.Set(false)

		if len(planConsumers) == 1 && planConsumers[0] == "*" {
			opts.Attributes.RemoteStateSharing = value.Set(true)
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
	_, err := r.ClientV2.Workspace.UpdateWorkspace(ctx, plan.Id.ValueString(), &opts)
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
			tags := make([]schemas.Tag, len(tagsToAdd))
			for i, tag := range tagsToAdd {
				tags[i] = schemas.Tag{ID: tag}
			}
			err = r.ClientV2.Workspace.AddWorkspaceTags(ctx, plan.Id.ValueString(), tags)
			if err != nil {
				resp.Diagnostics.AddError("Error adding tags to workspace", err.Error())
			}
		}

		if len(tagsToRemove) > 0 {
			tags := make([]schemas.Tag, len(tagsToRemove))
			for i, tag := range tagsToRemove {
				tags[i] = schemas.Tag{ID: tag}
			}
			err = r.ClientV2.Workspace.DeleteWorkspaceTags(ctx, plan.Id.ValueString(), tags)
			if err != nil {
				if !errors.Is(err, client.ErrNotFound) {
					// Tag resources may be removed by terraform earlier, ignore error if tag is not found.
					// This should be improved in the API: if the relationships are already missing
					// the server must return a successful response.
					resp.Diagnostics.AddError("Error removing tags from workspace", err.Error())
				}
			}
		}
	}

	if !plan.ProviderConfiguration.Equal(state.ProviderConfiguration) {
		expectedLinks := make(map[string]schemas.ProviderConfigurationLinkRequest)
		if !plan.ProviderConfiguration.IsNull() {
			var pcfgs []providerConfigurationModel
			resp.Diagnostics.Append(plan.ProviderConfiguration.ElementsAs(ctx, &pcfgs, false)...)

			for _, pcfg := range pcfgs {
				mapID := pcfg.ID.ValueString()
				pcfgReq := schemas.ProviderConfigurationLinkRequest{
					Relationships: schemas.ProviderConfigurationLinkRelationshipsRequest{
						ProviderConfiguration: value.Set(schemas.ProviderConfiguration{ID: pcfg.ID.ValueString()}),
					},
				}
				if !pcfg.Alias.IsUnknown() && len(pcfg.Alias.ValueString()) > 0 {
					pcfgReq.Attributes.Alias = value.Set(pcfg.Alias.ValueString())
					mapID = mapID + pcfg.Alias.ValueString()
				}
				expectedLinks[mapID] = pcfgReq
			}
		}

		currentLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.ClientV2, plan.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
		} else {
			for _, currentLink := range currentLinks {
				mapID := currentLink.Relationships.ProviderConfiguration.ID
				if currentLink.Attributes.Alias != nil {
					mapID = mapID + *currentLink.Attributes.Alias
				}
				if _, ok := expectedLinks[mapID]; ok {
					delete(expectedLinks, mapID)
				} else {
					err = r.ClientV2.ProviderConfigurationLink.DeleteProviderConfigurationWorkspaceLink(
						ctx,
						currentLink.ID,
					)
					if err != nil {
						resp.Diagnostics.AddError("Error deleting provider configuration link", err.Error())
					}
				}
			}

			for _, link := range expectedLinks {
				_, err = r.ClientV2.ProviderConfigurationLink.CreateProviderConfigurationLink(
					ctx,
					plan.Id.ValueString(),
					&link,
				)
				if err != nil {
					resp.Diagnostics.AddError("Error creating provider configuration link", err.Error())
				}
			}
		}
	}

	if !plan.SSHKeyID.Equal(state.SSHKeyID) {
		if plan.SSHKeyID.IsNull() {
			err = r.ClientV2.Misc.DeleteWorkspaceSshKeyLink(ctx, plan.Id.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error deleting SSH key link", err.Error())
			}
		} else {
			_, err = r.ClientV2.Misc.CreateWorkspaceSshKeyLink(
				ctx, plan.Id.ValueString(), &schemas.WorkspaceSSHKeyLinkRequest{
					SshKey: plan.SSHKeyID.ValueString(),
				},
			)
			if err != nil {
				resp.Diagnostics.AddError("Error creating SSH key link", err.Error())
			}
		}
	}

	if len(consumersToAdd) > 0 {
		c := make([]schemas.Workspace, len(consumersToAdd))
		for i, consumer := range consumersToAdd {
			c[i] = schemas.Workspace{ID: consumer}
		}
		err = r.ClientV2.Workspace.AddRemoteStateConsumers(ctx, plan.Id.ValueString(), c)
		if err != nil {
			resp.Diagnostics.AddError("Error adding remote state consumers", err.Error())
		}
	}
	if len(consumersToRemove) > 0 {
		c := make([]schemas.Workspace, len(consumersToRemove))
		for i, consumer := range consumersToRemove {
			c[i] = schemas.Workspace{ID: consumer}
		}
		err = r.ClientV2.Workspace.DeleteRemoteStateConsumers(ctx, plan.Id.ValueString(), c)
		if err != nil {
			resp.Diagnostics.AddError("Error removing remote state consumers", err.Error())
		}
	}

	// Get refreshed resource state from API
	ws, err := r.ClientV2.Workspace.GetWorkspace(
		ctx, plan.Id.ValueString(), &workspace.GetWorkspaceOptions{
			Include: []string{"created-by"},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.ClientV2, ws.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
	}

	stateConsumers, err := getRemoteStateConsumers(ctx, r.ClientV2, ws.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving remote state consumers", err.Error())
	}

	result, diags := workspaceResourceModelFromAPI(ctx, ws, pcfgLinks, stateConsumers, &plan)
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

	err := r.ClientV2.Workspace.DeleteWorkspace(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting workspace", err.Error())
		return
	}
}

func (r *workspaceResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
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
		if !operations.ValueBool() && executionMode.ValueString() == "remote" ||
			operations.ValueBool() && executionMode.ValueString() == "local" {
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
			resp.Diagnostics.Append(
				resp.Plan.SetAttribute(
					ctx,
					path.Root("execution_mode"),
					"local",
				)...,
			)
		} else {
			resp.Diagnostics.Append(
				resp.Plan.SetAttribute(
					ctx,
					path.Root("execution_mode"),
					"remote",
				)...,
			)
		}
	} else {
		// In all other cases, `execution_mode` dictates the value for `operations`, even when left default.
		if executionMode.ValueString() == "local" {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("operations"), false)...)
		} else {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("operations"), true)...)
		}
	}
}

// ImportState handles importing existing resources into Terraform state.
//
// In addition to default importing by resource ID,
// it is also possible to import the workspace by environment and workspace name
// in the format '<environment>/<workspace>'.
func (r *workspaceResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	if !strings.Contains(req.ID, "/") {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}

	parts := strings.SplitN(req.ID, "/", 2)
	envName := strings.TrimSpace(parts[0])
	wsName := strings.TrimSpace(parts[1])

	if envName == "" || wsName == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected '<environment>/<workspace>' with both parts non-empty.",
		)
		return
	}

	envs, err := r.ClientV2.Environment.ListEnvironments(
		ctx,
		&environment.ListEnvironmentsOptions{
			Filter: map[string]string{"name": envName},
			Fields: map[string]any{"environments": ""},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Import failed", fmt.Sprintf("Error listing environments: %s", err.Error()))
		return
	}
	if len(envs) != 1 {
		resp.Diagnostics.AddError(
			"Import failed",
			fmt.Sprintf("Expected exactly one environment with name %q, got %d.", envName, len(envs)),
		)
		return
	}

	wss, err := r.ClientV2.Workspace.GetWorkspaces(
		ctx, &workspace.GetWorkspacesOptions{
			Filter: map[string]string{"name": wsName, "environment": envs[0].ID},
			Fields: map[string]any{"workspaces": ""},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Import failed", fmt.Sprintf("Error listing workspaces: %s", err.Error()))
		return
	}
	if len(wss) != 1 {
		resp.Diagnostics.AddError(
			"Import failed",
			fmt.Sprintf("Expected exactly one workspace with name %q, got %d.", wsName, len(wss)),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), wss[0].ID)...)
}

func (r *workspaceResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   workspaceResourceSchemaV0(ctx),
			StateUpgrader: upgradeWorkspaceResourceStateV0toV4(r.ClientV2),
		},
		1: {
			PriorSchema:   workspaceResourceSchemaV1(ctx),
			StateUpgrader: upgradeWorkspaceResourceStateV1toV4(r.ClientV2),
		},
		2: {
			PriorSchema:   workspaceResourceSchemaV2(ctx),
			StateUpgrader: upgradeWorkspaceResourceStateV2toV4(r.ClientV2),
		},
		3: {
			PriorSchema:   workspaceResourceSchemaV3(ctx),
			StateUpgrader: upgradeWorkspaceResourceStateV3toV4(r.ClientV2),
		},
	}
}

func getProviderConfigurationWorkspaceLinks(
	ctx context.Context, scalrClient *scalrV2.Client, workspaceId string,
) (workspaceLinks []*schemas.ProviderConfigurationLink, err error) {
	for link, err := range scalrClient.ProviderConfigurationLink.ListProviderConfigurationLinksIter(
		ctx, workspaceId, nil,
	) {
		if err != nil {
			return nil, err
		}

		if link.Relationships.Workspace != nil {
			workspaceLinks = append(workspaceLinks, &link)
		}
	}

	return
}

func getRemoteStateConsumers(
	ctx context.Context, scalrClient *scalrV2.Client, workspaceId string,
) (consumers []string, err error) {
	for c, err := range scalrClient.Workspace.ListRemoteStateConsumersIter(ctx, workspaceId, nil) {
		if err != nil {
			return nil, err
		}

		consumers = append(consumers, c.ID)
	}
	return
}
