package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/scalr/go-scalr"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceScalrWorkspace() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of workspaces in Scalr. Create, update and destroy.",
		CreateContext: resourceScalrWorkspaceCreate,
		ReadContext:   resourceScalrWorkspaceRead,
		UpdateContext: resourceScalrWorkspaceUpdate,
		DeleteContext: resourceScalrWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 4,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceScalrWorkspaceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrWorkspaceStateUpgradeV0,
				Version: 0,
			},
			{
				Type:    resourceScalrWorkspaceResourceV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrWorkspaceStateUpgradeV1,
				Version: 1,
			},
			{
				Type:    resourceScalrWorkspaceResourceV2().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrWorkspaceStateUpgradeV2,
				Version: 2,
			},
			{
				Type:    resourceScalrWorkspaceResourceV3().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrWorkspaceStateUpgradeV3,
				Version: 3,
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the workspace.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"environment_id": {
				Description: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"vcs_provider_id": {
				Description:   "ID of VCS provider - required if vcs-repo present and vice versa, in the format `vcs-<RANDOM STRING>`.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"module_version_id"},
				RequiredWith:  []string{"vcs_repo"},
			},
			"module_version_id": {
				Description:   "The identifier of a module version in the format `modver-<RANDOM STRING>`. This attribute conflicts with `vcs_provider_id` and `vcs_repo` attributes.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"vcs_provider_id", "vcs_repo"},
			},
			"agent_pool_id": {
				Description: "The identifier of an agent pool in the format `apool-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"auto_apply": {
				Description: "Set (true/false) to configure if `terraform apply` should automatically run when `terraform plan` ends without error. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"force_latest_run": {
				Description: "Set (true/false) to configure if latest new run will be automatically raised in priority. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"deletion_protection_enabled": {
				Description: "Indicates if the workspace has the protection from an accidental state lost. If enabled and the workspace has resource, the deletion will not be allowed. Default `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},

			"var_files": {
				Description: "A list of paths to the `.tfvars` file(s) to be used as part of the workspace configuration.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},

			"operations": {
				Description: "Set (true/false) to configure workspace remote execution. When `false` workspace is only used to store state. Defaults to `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Deprecated:  "The attribute `operations` is deprecated. Use `execution_mode` instead",
			},

			"execution_mode": {
				Description: "Which execution mode to use. Valid values are `remote` and `local`. When set to `local`, the workspace will be used for state storage only. Defaults to `remote` (not set, backend default is used).",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.WorkspaceExecutionModeRemote),
						string(scalr.WorkspaceExecutionModeLocal),
					},
					false,
				),
			},

			"terraform_version": {
				Description: "The version of Terraform to use for this workspace. Defaults to the latest available version.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"terragrunt_version": {
				Description: "The version of Terragrunt the workspace performs runs on.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"terragrunt_use_run_all": {
				Description: "Indicates whether the workspace uses `terragrunt run-all`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},

			"iac_platform": {
				Description: "The IaC platform to use for this workspace. Valid values are `terraform` and `opentofu`. Defaults to `terraform`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     string(scalr.WorkspaceIaCPlatformTerraform),
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{
							string(scalr.WorkspaceIaCPlatformTerraform),
							string(scalr.WorkspaceIaCPlatformOpenTofu),
						},
						false,
					),
				),
			},

			"working_directory": {
				Description: "A relative path that Terraform will be run in. Defaults to the root of the repository `\"\"`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},

			"hooks": {
				Description: "Settings for the workspaces custom hooks.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pre_init": {
							Description: "Action that will be called before the init phase.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},

						"pre_plan": {
							Description: "Action that will be called before the plan phase.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},

						"post_plan": {
							Description: "Action that will be called after plan phase.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},

						"pre_apply": {
							Description: "Action that will be called before apply phase.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},

						"post_apply": {
							Description: "Action that will be called after apply phase.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},
					},
				},
			},

			"has_resources": {
				Description: "The presence of active terraform resources in the current state version.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"auto_queue_runs": {
				Description: "Indicates if runs have to be queued automatically when a new configuration version is uploaded. Supported values are `skip_first`, `always`, `never`:" +
					"\n  * `skip_first` - after the very first configuration version is uploaded into the workspace the run will not be triggered. But the following configurations will do. This is the default behavior." +
					"\n  * `always` - runs will be triggered automatically on every upload of the configuration version." +
					"\n  * `never` - configuration versions are uploaded into the workspace, but runs will not be triggered.",
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.AutoQueueRunsModeSkipFirst),
						string(scalr.AutoQueueRunsModeAlways),
						string(scalr.AutoQueueRunsModeNever),
					},
					false,
				),
				Computed: true,
			},

			"vcs_repo": {
				Description:   "Settings for the workspace's VCS repository.",
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				ConflictsWith: []string{"module_version_id"},
				RequiredWith:  []string{"vcs_provider_id"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "A reference to your VCS repository in the format `:org/:repo`, it refers to the organization and repository in your VCS provider.",
							Type:        schema.TypeString,
							Required:    true,
						},

						"branch": {
							Description: "The repository branch where Terraform will be run from. If omitted, the repository default branch will be used.",
							Type:        schema.TypeString,
							Optional:    true,
						},

						"path": {
							Description: "The repository subdirectory that Terraform will execute from. If omitted or submitted as an empty string, this defaults to the repository's root.",
							Type:        schema.TypeString,
							Default:     "",
							Optional:    true,
							Deprecated:  "The attribute `vcs-repo.path` is deprecated. Use working-directory and trigger-prefixes instead.",
						},

						"trigger_prefixes": {
							Description:   "List of paths (relative to `path`), whose changes will trigger a run for the workspace using this binding when the CV is created. Conflicts with `trigger_patterns`. If `trigger_prefixes` and `trigger_patterns` are omitted, any change in `path` will trigger a new run.",
							Type:          schema.TypeList,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"vcs_repo.0.trigger_patterns"},
						},
						"trigger_patterns": {
							Description:   "The gitignore-style patterns for files, whose changes will trigger a run for the workspace using this binding when the CV is created. Conflicts with `trigger_prefixes`. If `trigger_prefixes` and `trigger_patterns` are omitted, any change in `path` will trigger a new run.",
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"vcs_repo.0.trigger_prefixes"},
						},

						"dry_runs_enabled": {
							Description: "Set (true/false) to configure the VCS driven dry runs should run when pull request to configuration versions branch created. Default `true`.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"ingress_submodules": {
							Description: "Designates whether to clone git submodules of the VCS repository.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},

			"created_by": {
				Description: "Details of the user that created the workspace.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Description: "Username of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"email": {
							Description: "Email address of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"full_name": {
							Description: "Full name of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"run_operation_timeout": {
				Description: "The number of minutes run operation can be executed before termination. Defaults to `0` (not set, backend default is used).",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"type": {
				Description: "The type of the Scalr Workspace environment, available options: `production`, `staging`, `testing`, `development`, `unmapped`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.WorkspaceEnvironmentTypeProduction),
						string(scalr.WorkspaceEnvironmentTypeStaging),
						string(scalr.WorkspaceEnvironmentTypeTesting),
						string(scalr.WorkspaceEnvironmentTypeDevelopment),
						string(scalr.WorkspaceEnvironmentTypeUnmapped),
					},
					false,
				),
			},
			"provider_configuration": {
				Description: "Provider configurations used in workspace runs.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The identifier of provider configuration.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"alias": {
							Description: "The alias of provider configuration.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"ssh_key_id": {
				Description: "The identifier of the SSH key to use for the workspace.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tag_ids": {
				Description: "List of tag IDs associated with the workspace.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func parseTriggerPrefixDefinitions(vcsRepo map[string]interface{}) ([]string, error) {
	triggerPrefixes := make([]string, 0)

	triggerPrefixIds := vcsRepo["trigger_prefixes"].([]interface{})
	err := ValidateIDsDefinitions(triggerPrefixIds)
	if err != nil {
		return nil, fmt.Errorf("Got error during parsing trigger prefixes: %s", err.Error())
	}

	for _, triggerPrefixId := range triggerPrefixIds {
		triggerPrefixes = append(triggerPrefixes, triggerPrefixId.(string))
	}

	return triggerPrefixes, nil
}

func resourceScalrWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get the name, environment_id and vcs_provider_id.
	name := d.Get("name").(string)
	environmentID := d.Get("environment_id").(string)

	// Create a new options struct.
	options := scalr.WorkspaceCreateOptions{
		Name:                      ptr(name),
		AutoApply:                 ptr(d.Get("auto_apply").(bool)),
		ForceLatestRun:            ptr(d.Get("force_latest_run").(bool)),
		DeletionProtectionEnabled: ptr(d.Get("deletion_protection_enabled").(bool)),
		Environment:               &scalr.Environment{ID: environmentID},
		Hooks:                     &scalr.HooksOptions{},
	}

	// Process all configured options.
	if operations, ok := d.GetOk("operations"); ok {
		options.Operations = ptr(operations.(bool))
	}
	if terragruntVersion, ok := d.GetOk("terragrunt_version"); ok {
		options.TerragruntVersion = ptr(terragruntVersion.(string))
	}
	if terragruntUseRunAll, ok := d.GetOk("terragrunt_use_run_all"); ok {
		options.TerragruntUseRunAll = ptr(terragruntUseRunAll.(bool))
	}

	if executionMode, ok := d.GetOk("execution_mode"); ok {
		options.ExecutionMode = ptr(
			scalr.WorkspaceExecutionMode(executionMode.(string)),
		)
	}

	if autoQueueRunsI, ok := d.GetOk("auto_queue_runs"); ok {
		options.AutoQueueRuns = ptr(
			scalr.WorkspaceAutoQueueRuns(autoQueueRunsI.(string)),
		)
	}

	if workspaceEnvironmentTypeI, ok := d.GetOk("type"); ok {
		options.EnvironmentType = ptr(
			scalr.WorkspaceEnvironmentType(workspaceEnvironmentTypeI.(string)),
		)
	}

	if tfVersion, ok := d.GetOk("terraform_version"); ok {
		options.TerraformVersion = ptr(tfVersion.(string))
	}

	if iacPlatform, ok := d.GetOk("iac_platform"); ok {
		options.IacPlatform = ptr(scalr.WorkspaceIaCPlatform(iacPlatform.(string)))
	}

	if workingDir, ok := d.GetOk("working_directory"); ok {
		options.WorkingDirectory = ptr(workingDir.(string))
	}

	if runOperationTimeout, ok := d.GetOk("run_operation_timeout"); ok {
		options.RunOperationTimeout = ptr(runOperationTimeout.(int))
	}

	if v, ok := d.GetOk("module_version_id"); ok {
		options.ModuleVersion = &scalr.ModuleVersion{ID: v.(string)}
	}

	if vcsProviderID, ok := d.GetOk("vcs_provider_id"); ok {
		options.VcsProvider = &scalr.VcsProvider{
			ID: vcsProviderID.(string),
		}
	}

	if agentPoolID, ok := d.GetOk("agent_pool_id"); ok {
		options.AgentPool = &scalr.AgentPool{
			ID: agentPoolID.(string),
		}
	}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})
		triggerPrefixes, err := parseTriggerPrefixDefinitions(vcsRepo)
		if err != nil {
			return diag.FromErr(err)
		}

		options.VCSRepo = &scalr.WorkspaceVCSRepoOptions{
			Identifier:        ptr(vcsRepo["identifier"].(string)),
			Path:              ptr(vcsRepo["path"].(string)),
			TriggerPrefixes:   &triggerPrefixes,
			TriggerPatterns:   ptr(vcsRepo["trigger_patterns"].(string)),
			DryRunsEnabled:    ptr(vcsRepo["dry_runs_enabled"].(bool)),
			IngressSubmodules: ptr(vcsRepo["ingress_submodules"].(bool)),
		}

		// Only set the branch if one is configured.
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			options.VCSRepo.Branch = ptr(branch)
		}
	}

	// Get and assert the hooks
	if v, ok := d.GetOk("hooks"); ok {
		if _, ok := v.([]interface{})[0].(map[string]interface{}); ok {
			hooks := v.([]interface{})[0].(map[string]interface{})

			options.Hooks = &scalr.HooksOptions{
				PreInit:   ptr(hooks["pre_init"].(string)),
				PrePlan:   ptr(hooks["pre_plan"].(string)),
				PostPlan:  ptr(hooks["post_plan"].(string)),
				PreApply:  ptr(hooks["pre_apply"].(string)),
				PostApply: ptr(hooks["post_apply"].(string)),
			}
		}
	}

	if v, ok := d.GetOk("var_files"); ok {
		vfiles := v.([]interface{})
		varFiles := make([]string, 0)
		for _, varFile := range vfiles {
			varFiles = append(varFiles, varFile.(string))
		}
		options.VarFiles = varFiles
	}

	if tagIDs, ok := d.GetOk("tag_ids"); ok {
		tagIDsList := tagIDs.(*schema.Set).List()
		tags := make([]*scalr.Tag, len(tagIDsList))
		for i, id := range tagIDsList {
			tags[i] = &scalr.Tag{ID: id.(string)}
		}
		options.Tags = tags
	}

	log.Printf("[DEBUG] Create workspace %s for environment: %s", name, environmentID)
	workspace, err := scalrClient.Workspaces.Create(ctx, options)
	if err != nil {
		return diag.Errorf(
			"Error creating workspace %s for environment %s: %v", name, environmentID, err)
	}
	d.SetId(workspace.ID)

	if providerConfigurationsI, ok := d.GetOk("provider_configuration"); ok {
		for _, v := range providerConfigurationsI.(*schema.Set).List() {
			pcfg := v.(map[string]interface{})
			createLinkOption := scalr.ProviderConfigurationLinkCreateOptions{
				ProviderConfiguration: &scalr.ProviderConfiguration{ID: pcfg["id"].(string)},
			}
			if alias, ok := pcfg["alias"]; ok && len(alias.(string)) > 0 {
				createLinkOption.Alias = ptr(alias.(string))
			}
			_, err := scalrClient.ProviderConfigurationLinks.Create(
				ctx, workspace.ID, createLinkOption,
			)
			if err != nil {
				return diag.Errorf(
					"Error creating workspace %s provider configuration link: %v", name, err)
			}
		}
	}

	if sshKeyID, ok := d.GetOk("ssh_key_id"); ok {
		_, err := scalrClient.SSHKeysLinks.Create(ctx, workspace.ID, sshKeyID.(string))
		if err != nil {
			return diag.Errorf("Error creating SSH key link for workspace %s: %v", name, err)
		}
	}

	return resourceScalrWorkspaceRead(ctx, d, meta)
}

func resourceScalrWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	log.Printf("[DEBUG] Read configuration of workspace: %s", id)
	workspace, err := scalrClient.Workspaces.ReadByID(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Workspace %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading configuration of workspace %s: %v", id, err)
	}

	// Update the config.
	_ = d.Set("name", workspace.Name)
	_ = d.Set("auto_apply", workspace.AutoApply)
	_ = d.Set("force_latest_run", workspace.ForceLatestRun)
	_ = d.Set("deletion_protection_enabled", workspace.DeletionProtectionEnabled)
	_ = d.Set("operations", workspace.Operations)
	_ = d.Set("execution_mode", workspace.ExecutionMode)
	_ = d.Set("terraform_version", workspace.TerraformVersion)
	_ = d.Set("iac_platform", workspace.IaCPlatform)
	_ = d.Set("working_directory", workspace.WorkingDirectory)
	_ = d.Set("environment_id", workspace.Environment.ID)
	_ = d.Set("has_resources", workspace.HasResources)
	_ = d.Set("auto_queue_runs", workspace.AutoQueueRuns)
	_ = d.Set("type", workspace.EnvironmentType)
	_ = d.Set("var_files", workspace.VarFiles)
	_ = d.Set("terragrunt_version", workspace.TerraformVersion)
	_ = d.Set("terragrunt_use_run_all", workspace.TerragruntUseRunAll)

	if workspace.RunOperationTimeout != nil {
		_ = d.Set("run_operation_timeout", &workspace.RunOperationTimeout)
	}

	if workspace.VcsProvider != nil {
		_ = d.Set("vcs_provider_id", workspace.VcsProvider.ID)
	}

	if workspace.SSHKey != nil {
		_ = d.Set("ssh_key_id", workspace.SSHKey.ID)
	}

	if workspace.AgentPool != nil {
		_ = d.Set("agent_pool_id", workspace.AgentPool.ID)
	} else {
		_ = d.Set("agent_pool_id", "")
	}

	var mv string
	if workspace.ModuleVersion != nil {
		mv = workspace.ModuleVersion.ID
	}
	_ = d.Set("module_version_id", mv)

	var createdBy []interface{}
	if workspace.CreatedBy != nil {
		createdBy = append(createdBy, map[string]interface{}{
			"username":  workspace.CreatedBy.Username,
			"email":     workspace.CreatedBy.Email,
			"full_name": workspace.CreatedBy.FullName,
		})
	}
	_ = d.Set("created_by", createdBy)

	var vcsRepo []interface{}
	if workspace.VCSRepo != nil {
		vcsRepo = append(vcsRepo, map[string]interface{}{
			"branch":             workspace.VCSRepo.Branch,
			"identifier":         workspace.VCSRepo.Identifier,
			"path":               workspace.VCSRepo.Path,
			"trigger_prefixes":   workspace.VCSRepo.TriggerPrefixes,
			"trigger_patterns":   workspace.VCSRepo.TriggerPatterns,
			"dry_runs_enabled":   workspace.VCSRepo.DryRunsEnabled,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
		})
	}
	_ = d.Set("vcs_repo", vcsRepo)

	var hooks []interface{}
	if workspace.Hooks != nil {
		hooks = append(hooks, map[string]interface{}{
			"pre_init":   workspace.Hooks.PreInit,
			"pre_plan":   workspace.Hooks.PrePlan,
			"post_plan":  workspace.Hooks.PostPlan,
			"pre_apply":  workspace.Hooks.PreApply,
			"post_apply": workspace.Hooks.PostApply,
		})
	} else if _, ok := d.GetOk("hooks"); ok {
		hooks = append(hooks, map[string]interface{}{
			"pre_init":   "",
			"pre_plan":   "",
			"post_plan":  "",
			"pre_apply":  "",
			"post_apply": "",
		})
	}
	_ = d.Set("hooks", hooks)

	providerConfigurationLinks, err := getProviderConfigurationWorkspaceLinks(ctx, scalrClient, id)
	if err != nil {
		return diag.Errorf("Error reading provider configuration links of workspace %s: %v", id, err)
	}
	var providerConfigurations []map[string]interface{}
	for _, link := range providerConfigurationLinks {
		providerConfigurations = append(providerConfigurations, map[string]interface{}{
			"id":    link.ProviderConfiguration.ID,
			"alias": link.Alias,
		})
	}
	_ = d.Set("provider_configuration", providerConfigurations)

	var tagIDs []string
	if len(workspace.Tags) != 0 {
		for _, tag := range workspace.Tags {
			tagIDs = append(tagIDs, tag.ID)
		}
	}
	_ = d.Set("tag_ids", tagIDs)

	return nil
}

func resourceScalrWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("auto_apply") || d.HasChange("auto_queue_runs") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") || d.HasChange("force_latest_run") ||
		d.HasChange("vcs_repo") || d.HasChange("operations") || d.HasChange("execution_mode") ||
		d.HasChange("vcs_provider_id") || d.HasChange("agent_pool_id") || d.HasChange("deletion_protection_enabled") ||
		d.HasChange("hooks") || d.HasChange("module_version_id") || d.HasChange("var_files") ||
		d.HasChange("run_operation_timeout") || d.HasChange("iac_platform") ||
		d.HasChange("type") || d.HasChange("terragrunt_version") || d.HasChange("terragrunt_use_run_all") {
		// Create a new options struct.
		options := scalr.WorkspaceUpdateOptions{
			Name:                      ptr(d.Get("name").(string)),
			AutoApply:                 ptr(d.Get("auto_apply").(bool)),
			ForceLatestRun:            ptr(d.Get("force_latest_run").(bool)),
			DeletionProtectionEnabled: ptr(d.Get("deletion_protection_enabled").(bool)),
			Hooks: &scalr.HooksOptions{
				PreInit:   ptr(""),
				PrePlan:   ptr(""),
				PostPlan:  ptr(""),
				PreApply:  ptr(""),
				PostApply: ptr(""),
			},
		}

		// Process all configured options.
		if operations, ok := d.GetOk("operations"); ok {
			options.Operations = ptr(operations.(bool))
		}

		if executionMode, ok := d.GetOk("execution_mode"); ok {
			options.ExecutionMode = ptr(
				scalr.WorkspaceExecutionMode(executionMode.(string)),
			)
		}
		if terragruntVersion, ok := d.GetOk("terragrunt_version"); ok {
			options.TerragruntVersion = ptr(terragruntVersion.(string))
		}
		if terragruntUseRunAll, ok := d.GetOkExists("terragrunt_use_run_all"); ok { //nolint:staticcheck
			options.TerragruntUseRunAll = ptr(terragruntUseRunAll.(bool))
		}

		if autoQueueRunsI, ok := d.GetOk("auto_queue_runs"); ok {
			options.AutoQueueRuns = ptr(
				scalr.WorkspaceAutoQueueRuns(autoQueueRunsI.(string)),
			)
		}

		if workspaceEnvironmentTypeI, ok := d.GetOk("type"); ok {
			options.EnvironmentType = ptr(
				scalr.WorkspaceEnvironmentType(workspaceEnvironmentTypeI.(string)),
			)
		}

		if tfVersion, ok := d.GetOk("terraform_version"); ok {
			options.TerraformVersion = ptr(tfVersion.(string))
		}

		if iacPlatform, ok := d.GetOk("iac_platform"); ok {
			options.IacPlatform = ptr(scalr.WorkspaceIaCPlatform(iacPlatform.(string)))
		}

		if v, ok := d.Get("var_files").([]interface{}); ok {
			varFiles := make([]string, 0)
			for _, varFile := range v {
				varFiles = append(varFiles, varFile.(string))
			}
			options.VarFiles = varFiles
		}

		options.WorkingDirectory = ptr(d.Get("working_directory").(string))

		if runOperationTimeout, ok := d.GetOk("run_operation_timeout"); ok {
			options.RunOperationTimeout = ptr(runOperationTimeout.(int))
		}

		if vcsProviderId, ok := d.GetOk("vcs_provider_id"); ok {
			options.VcsProvider = &scalr.VcsProvider{
				ID: vcsProviderId.(string),
			}
		}

		if agentPoolID, ok := d.GetOk("agent_pool_id"); ok {
			options.AgentPool = &scalr.AgentPool{
				ID: agentPoolID.(string),
			}
		}

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			vcsRepo := v.([]interface{})[0].(map[string]interface{})
			triggerPrefixes, err := parseTriggerPrefixDefinitions(vcsRepo)
			if err != nil {
				return diag.FromErr(err)
			}

			options.VCSRepo = &scalr.WorkspaceVCSRepoOptions{
				Identifier:        ptr(vcsRepo["identifier"].(string)),
				Branch:            ptr(vcsRepo["branch"].(string)),
				Path:              ptr(vcsRepo["path"].(string)),
				TriggerPrefixes:   &triggerPrefixes,
				TriggerPatterns:   ptr(vcsRepo["trigger_patterns"].(string)),
				DryRunsEnabled:    ptr(vcsRepo["dry_runs_enabled"].(bool)),
				IngressSubmodules: ptr(vcsRepo["ingress_submodules"].(bool)),
			}
		}

		// Get and assert the hooks
		if v, ok := d.GetOk("hooks"); ok {
			if _, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				hooks := v.([]interface{})[0].(map[string]interface{})

				options.Hooks = &scalr.HooksOptions{
					PreInit:   ptr(hooks["pre_init"].(string)),
					PrePlan:   ptr(hooks["pre_plan"].(string)),
					PostPlan:  ptr(hooks["post_plan"].(string)),
					PreApply:  ptr(hooks["pre_apply"].(string)),
					PostApply: ptr(hooks["post_apply"].(string)),
				}
			}
		}

		if v, ok := d.GetOk("module_version_id"); ok {
			options.ModuleVersion = &scalr.ModuleVersion{
				ID: v.(string),
			}
		}

		log.Printf("[DEBUG] Update workspace %s", id)
		_, err := scalrClient.Workspaces.Update(ctx, id, options)
		if err != nil {
			return diag.Errorf(
				"Error updating workspace %s: %v", id, err)
		}
	}
	if d.HasChange("ssh_key_id") {
		oldSSHKeyID, newSSHKeyID := d.GetChange("ssh_key_id")

		if oldSSHKeyID != "" && newSSHKeyID == "" {
			err := scalrClient.SSHKeysLinks.Delete(ctx, id)
			if err != nil {
				return diag.Errorf("Error removing SSH key link for workspace %s: %v", id, err)
			}
		}

		if newSSHKeyID != "" {
			_, err := scalrClient.SSHKeysLinks.Create(ctx, id, newSSHKeyID.(string))
			if err != nil {
				return diag.Errorf("Error creating SSH key link for workspace %s: %v", id, err)
			}
		}
	}

	if d.HasChange("provider_configuration") {

		expectedLinks := make(map[string]scalr.ProviderConfigurationLinkCreateOptions)
		if providerConfigurationI, ok := d.GetOk("provider_configuration"); ok {
			for _, v := range providerConfigurationI.(*schema.Set).List() {
				configLink := v.(map[string]interface{})
				mapID := configLink["id"].(string)
				linkCreateOption := scalr.ProviderConfigurationLinkCreateOptions{
					ProviderConfiguration: &scalr.ProviderConfiguration{ID: configLink["id"].(string)},
				}
				if v, ok := configLink["alias"]; ok && len(v.(string)) > 0 {
					linkCreateOption.Alias = ptr(v.(string))
					mapID = mapID + v.(string)
				}
				expectedLinks[mapID] = linkCreateOption

			}
		}

		currentLinks, err := getProviderConfigurationWorkspaceLinks(ctx, scalrClient, id)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, currentLink := range currentLinks {
			mapID := currentLink.ProviderConfiguration.ID + currentLink.Alias
			if _, ok := expectedLinks[mapID]; ok {
				delete(expectedLinks, mapID)
			} else {
				err = scalrClient.ProviderConfigurationLinks.Delete(ctx, currentLink.ID)
				if err != nil {
					return diag.Errorf(
						"Error removing provider configuration link in workspace %s: %v", id, err)
				}
			}
		}
		for _, createOption := range expectedLinks {
			_, err = scalrClient.ProviderConfigurationLinks.Create(ctx, id, createOption)
			if err != nil {
				return diag.Errorf(
					"Error creating provider configuration link in workspace %s: %v", id, err)
			}

		}
	}

	if d.HasChange("tag_ids") {
		oldTags, newTags := d.GetChange("tag_ids")
		oldSet := oldTags.(*schema.Set)
		newSet := newTags.(*schema.Set)
		tagsToAdd := InterfaceArrToTagRelationArr(newSet.Difference(oldSet).List())
		tagsToDelete := InterfaceArrToTagRelationArr(oldSet.Difference(newSet).List())

		if len(tagsToAdd) > 0 {
			err := scalrClient.WorkspaceTags.Add(ctx, id, tagsToAdd)
			if err != nil {
				return diag.Errorf(
					"Error adding tags to workspace %s: %v", id, err)
			}
		}

		if len(tagsToDelete) > 0 {
			err := scalrClient.WorkspaceTags.Delete(ctx, id, tagsToDelete)
			if err != nil {
				return diag.Errorf(
					"Error deleting tags from workspace %s: %v", id, err)
			}
		}
	}

	return resourceScalrWorkspaceRead(ctx, d, meta)
}

func getProviderConfigurationWorkspaceLinks(
	ctx context.Context, scalrClient *scalr.Client, workspaceId string,
) (workspaceLinks []*scalr.ProviderConfigurationLink, err error) {
	linkListOption := scalr.ProviderConfigurationLinksListOptions{Include: "provider-configuration"}
	for {
		linksList, err := scalrClient.ProviderConfigurationLinks.List(ctx, workspaceId, linkListOption)

		if err != nil {
			return nil, fmt.Errorf("Error reading provider configuration links %s: %v", workspaceId, err)
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

func resourceScalrWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete workspace %s", id)
	err := scalrClient.Workspaces.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting workspace %s: %v", id, err)
	}

	return nil
}
