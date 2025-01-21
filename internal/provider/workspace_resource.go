package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &workspaceResource{}
	_ resource.ResourceWithConfigure        = &workspaceResource{}
	_ resource.ResourceWithConfigValidators = &workspaceResource{}
	_ resource.ResourceWithModifyPlan       = &workspaceResource{}
	_ resource.ResourceWithImportState      = &workspaceResource{}
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
	emptyStringList, _ := types.ListValueFrom(ctx, types.StringType, []string{})
	emptyStringSet, _ := types.SetValueFrom(ctx, types.StringType, []string{})

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of workspaces in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the workspace.",
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "ID of VCS provider - required if vcs-repo present and vice versa, in the format `vcs-<RANDOM STRING>`.",
				Optional:            true,
			},
			"module_version_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a module version in the format `modver-<RANDOM STRING>`. This attribute conflicts with `vcs_provider_id` and `vcs_repo` attributes.",
				Optional:            true,
			},
			"agent_pool_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of an agent pool in the format `apool-<RANDOM STRING>`.",
				Optional:            true,
			},
			"auto_apply": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure if `terraform apply` should automatically run when `terraform plan` ends without error. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"force_latest_run": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure if latest new run will be automatically raised in priority. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"deletion_protection_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the workspace has the protection from an accidental state lost. If enabled and the workspace has resource, the deletion will not be allowed. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"var_files": schema.ListAttribute{
				MarkdownDescription: "A list of paths to the `.tfvars` file(s) to be used as part of the workspace configuration.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(emptyStringList),
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
			"operations": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure workspace remote execution. When `false` workspace is only used to store state. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				DeprecationMessage:  "The attribute `operations` is deprecated. Use `execution_mode` instead",
			},
			"execution_mode": schema.StringAttribute{
				MarkdownDescription: "Which execution mode to use. Valid values are `remote` and `local`. When set to `local`, the workspace will be used for state storage only. Defaults to `remote`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(scalr.WorkspaceExecutionModeRemote)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.WorkspaceExecutionModeRemote),
						string(scalr.WorkspaceExecutionModeLocal),
					),
				},
			},
			"terraform_version": schema.StringAttribute{
				MarkdownDescription: "The version of Terraform to use for this workspace. Defaults to the latest available version.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"terragrunt_version": schema.StringAttribute{
				MarkdownDescription: "The version of Terragrunt the workspace performs runs on.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"terragrunt_use_run_all": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the workspace uses `terragrunt run-all`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"iac_platform": schema.StringAttribute{
				MarkdownDescription: "The IaC platform to use for this workspace. Valid values are `terraform` and `opentofu`. Defaults to `terraform`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(scalr.WorkspaceIaCPlatformTerraform)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.WorkspaceIaCPlatformTerraform),
						string(scalr.WorkspaceIaCPlatformOpenTofu),
					),
				},
			},
			"working_directory": schema.StringAttribute{
				MarkdownDescription: "A relative path that Terraform will be run in. Defaults to the root of the repository `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"has_resources": schema.BoolAttribute{
				MarkdownDescription: "The presence of active terraform resources in the current state version.",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_queue_runs": schema.StringAttribute{
				MarkdownDescription: "Indicates if runs have to be queued automatically when a new configuration version is uploaded. Supported values are `skip_first`, `always`, `never`:" +
					"\n  * `skip_first` - after the very first configuration version is uploaded into the workspace the run will not be triggered. But the following configurations will do. This is the default behavior." +
					"\n  * `always` - runs will be triggered automatically on every upload of the configuration version." +
					"\n  * `never` - configuration versions are uploaded into the workspace, but runs will not be triggered.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(scalr.AutoQueueRunsModeSkipFirst)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.AutoQueueRunsModeSkipFirst),
						string(scalr.AutoQueueRunsModeAlways),
						string(scalr.AutoQueueRunsModeNever),
					),
				},
			},
			"created_by": schema.ListAttribute{
				MarkdownDescription: "Details of the user that created the workspace.",
				ElementType:         userElementType,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"run_operation_timeout": schema.Int32Attribute{
				MarkdownDescription: "The number of minutes run operation can be executed before termination.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the Scalr Workspace environment, available options: `production`, `staging`, `testing`, `development`, `unmapped`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(scalr.WorkspaceEnvironmentTypeUnmapped)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.WorkspaceEnvironmentTypeProduction),
						string(scalr.WorkspaceEnvironmentTypeStaging),
						string(scalr.WorkspaceEnvironmentTypeTesting),
						string(scalr.WorkspaceEnvironmentTypeDevelopment),
						string(scalr.WorkspaceEnvironmentTypeUnmapped),
					),
				},
			},
			"ssh_key_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the SSH key to use for the workspace.",
				Optional:            true,
			},
			"tag_ids": schema.SetAttribute{
				MarkdownDescription: "List of tag IDs associated with the workspace.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(emptyStringSet),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"hooks": schema.ListNestedBlock{
				MarkdownDescription: "Settings for the workspaces custom hooks.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"pre_init": schema.StringAttribute{
							MarkdownDescription: "Action that will be called before the init phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"pre_plan": schema.StringAttribute{
							MarkdownDescription: "Action that will be called before the plan phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"post_plan": schema.StringAttribute{
							MarkdownDescription: "Action that will be called after plan phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"pre_apply": schema.StringAttribute{
							MarkdownDescription: "Action that will be called before apply phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"post_apply": schema.StringAttribute{
							MarkdownDescription: "Action that will be called after apply phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"vcs_repo": schema.ListNestedBlock{
				MarkdownDescription: "Settings for the workspace's VCS repository.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							MarkdownDescription: "A reference to your VCS repository in the format `:org/:repo`, it refers to the organization and repository in your VCS provider.",
							Required:            true,
						},
						"branch": schema.StringAttribute{
							MarkdownDescription: "The repository branch where Terraform will be run from. If omitted, the repository default branch will be used.",
							Optional:            true,
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The repository subdirectory that Terraform will execute from. If omitted or submitted as an empty string, this defaults to the repository's root.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
							DeprecationMessage:  "The attribute `vcs-repo.path` is deprecated. Use working-directory and trigger-prefixes instead.",
						},
						"trigger_prefixes": schema.ListAttribute{
							MarkdownDescription: "List of paths (relative to `path`), whose changes will trigger a run for the workspace using this binding when the CV is created. Conflicts with `trigger_patterns`. If `trigger_prefixes` and `trigger_patterns` are omitted, any change in `path` will trigger a new run.",
							ElementType:         types.StringType,
							Optional:            true,
						},
						"trigger_patterns": schema.StringAttribute{
							MarkdownDescription: "The gitignore-style patterns for files, whose changes will trigger a run for the workspace using this binding when the CV is created. Conflicts with `trigger_prefixes`. If `trigger_prefixes` and `trigger_patterns` are omitted, any change in `path` will trigger a new run.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("trigger_prefixes")),
							},
						},
						"dry_runs_enabled": schema.BoolAttribute{
							MarkdownDescription: "Set (true/false) to configure the VCS driven dry runs should run when pull request to configuration versions branch created. Default `true`.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
						},
						"ingress_submodules": schema.BoolAttribute{
							MarkdownDescription: "Designates whether to clone git submodules of the VCS repository.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"provider_configuration": schema.SetNestedBlock{
				MarkdownDescription: "Provider configurations used in workspace runs.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of provider configuration.",
							Required:            true,
						},
						"alias": schema.StringAttribute{
							MarkdownDescription: "The alias of provider configuration.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
					},
				},
			},
		},
	}
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
	var plan workspaceModel

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
		TerragruntUseRunAll:       plan.TerragruntUseRunAll.ValueBoolPointer(),
		VarFiles:                  varFiles,
		WorkingDirectory:          plan.WorkingDirectory.ValueStringPointer(),
		Environment: &scalr.Environment{
			ID: plan.EnvironmentID.ValueString(),
		},
	}

	if !plan.TerraformVersion.IsUnknown() && !plan.TerraformVersion.IsNull() {
		opts.TerraformVersion = plan.TerraformVersion.ValueStringPointer()
	}

	if !plan.TerragruntVersion.IsUnknown() && !plan.TerragruntVersion.IsNull() {
		opts.TerragruntVersion = plan.TerragruntVersion.ValueStringPointer()
	}

	if !plan.RunOperationTimeout.IsUnknown() && !plan.RunOperationTimeout.IsNull() {
		opts.RunOperationTimeout = ptr(int(plan.RunOperationTimeout.ValueInt32()))
	}

	if !plan.VCSProviderID.IsUnknown() && !plan.VCSProviderID.IsNull() {
		opts.VcsProvider = &scalr.VcsProvider{
			ID: plan.VCSProviderID.ValueString(),
		}
	}

	if !plan.ModuleVersion.IsUnknown() && !plan.ModuleVersion.IsNull() {
		opts.ModuleVersion = &scalr.ModuleVersion{
			ID: plan.ModuleVersion.ValueString(),
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

	// Get refreshed resource state from API
	workspace, err = r.Client.Workspaces.ReadByID(ctx, workspace.ID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.Client, workspace.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
		return
	}

	result, diags := workspaceModelFromAPI(ctx, workspace, pcfgLinks, &plan)
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
	var state workspaceModel
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
		return
	}

	result, diags := workspaceModelFromAPI(ctx, workspace, pcfgLinks, &state)
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
	var plan, state workspaceModel
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

	if !plan.TerragruntUseRunAll.Equal(state.TerragruntUseRunAll) {
		opts.TerragruntUseRunAll = plan.TerragruntUseRunAll.ValueBoolPointer()
	}

	if !plan.TerragruntVersion.Equal(state.TerragruntVersion) && !plan.TerragruntVersion.IsNull() {
		opts.TerragruntVersion = plan.TerragruntVersion.ValueStringPointer()
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

	if !plan.ModuleVersion.IsNull() {
		opts.ModuleVersion = &scalr.ModuleVersion{ID: plan.ModuleVersion.ValueString()}
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

	// Get refreshed resource state from API
	workspace, err := r.Client.Workspaces.ReadByID(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}

	pcfgLinks, err := getProviderConfigurationWorkspaceLinks(ctx, r.Client, workspace.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving provider configuration links", err.Error())
		return
	}

	result, diags := workspaceModelFromAPI(ctx, workspace, pcfgLinks, &plan)
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
	var state workspaceModel
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
		if operations.ValueBool() == false && executionMode.ValueString() == string(scalr.WorkspaceExecutionModeRemote) ||
			operations.ValueBool() == true && executionMode.ValueString() == string(scalr.WorkspaceExecutionModeLocal) {
			resp.Diagnostics.AddError(
				"Attributes `operations` and `execution_mode` are configured with conflicting values",
				"The attribute `operations` is deprecated. Use `execution_mode` instead",
			)
		}
	}

	if !operationsCfg.IsNull() && executionModeCfg.IsNull() {
		// When the `operations` is explicitly set in the configuration, and `execution_mode` is not -
		// this is the only case when it takes precedence.
		if operations.ValueBool() == false {
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
