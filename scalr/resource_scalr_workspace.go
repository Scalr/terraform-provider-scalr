package scalr

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	scalr "github.com/scalr/go-scalr"
)

func resourceScalrWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrWorkspaceCreate,
		ReadContext:   resourceScalrWorkspaceRead,
		UpdateContext: resourceScalrWorkspaceUpdate,
		DeleteContext: resourceScalrWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:     schema.TypeString,
				Required: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vcs_provider_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"module_version_id"},
				RequiredWith:  []string{"vcs_repo"},
			},
			"module_version_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"vcs_provider_id", "vcs_repo"},
			},
			"agent_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"auto_apply": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"force_latest_run": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"var_files": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"operations": {
				Type:       schema.TypeBool,
				Optional:   true,
				Computed:   true,
				Deprecated: "The attribute `operations` is deprecated. Use `execution-mode` instead",
			},

			"execution_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.WorkspaceExecutionModeRemote),
						string(scalr.WorkspaceExecutionModeLocal),
					},
					false,
				),
			},

			"terraform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"hooks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pre_init": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},

						"pre_plan": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},

						"post_plan": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},

						"pre_apply": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},

						"post_apply": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},

			"has_resources": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"auto_queue_runs": {
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
				Default: string(scalr.AutoQueueRunsModeSkipFirst),
			},

			"vcs_repo": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				ConflictsWith: []string{"module_version_id"},
				RequiredWith:  []string{"vcs_provider_id"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:     schema.TypeString,
							Required: true,
						},

						"branch": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"path": {
							Type:       schema.TypeString,
							Default:    "",
							Optional:   true,
							Deprecated: "The attribute `vcs-repo.path` is deprecated. Use working-directory and trigger-prefixes instead.",
						},

						"trigger_prefixes": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							Computed: true,
						},

						"dry_runs_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"ingress_submodules": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"created_by": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"full_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"run_operation_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"provider_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"alias": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tag_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		Name:           scalr.String(name),
		AutoApply:      scalr.Bool(d.Get("auto_apply").(bool)),
		ForceLatestRun: scalr.Bool(d.Get("force_latest_run").(bool)),
		Environment:    &scalr.Environment{ID: environmentID},
		Hooks:          &scalr.HooksOptions{},
	}

	// Process all configured options.
	if operations, ok := d.GetOk("operations"); ok {
		options.Operations = scalr.Bool(operations.(bool))
	}

	if executionMode, ok := d.GetOk("execution_mode"); ok {
		options.ExecutionMode = scalr.WorkspaceExecutionModePtr(
			scalr.WorkspaceExecutionMode(executionMode.(string)),
		)
	}

	if autoQueueRunsI, ok := d.GetOk("auto_queue_runs"); ok {
		options.AutoQueueRuns = scalr.AutoQueueRunsModePtr(
			scalr.WorkspaceAutoQueueRuns(autoQueueRunsI.(string)),
		)
	}

	if tfVersion, ok := d.GetOk("terraform_version"); ok {
		options.TerraformVersion = scalr.String(tfVersion.(string))
	}

	if workingDir, ok := d.GetOk("working_directory"); ok {
		options.WorkingDirectory = scalr.String(workingDir.(string))
	}

	if runOperationTimeout, ok := d.GetOk("run_operation_timeout"); ok {
		options.RunOperationTimeout = scalr.Int(runOperationTimeout.(int))
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
			Identifier:        scalr.String(vcsRepo["identifier"].(string)),
			Path:              scalr.String(vcsRepo["path"].(string)),
			TriggerPrefixes:   &triggerPrefixes,
			DryRunsEnabled:    scalr.Bool(vcsRepo["dry_runs_enabled"].(bool)),
			IngressSubmodules: scalr.Bool(vcsRepo["ingress_submodules"].(bool)),
		}

		// Only set the branch if one is configured.
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			options.VCSRepo.Branch = scalr.String(branch)
		}
	}

	// Get and assert the hooks
	if v, ok := d.GetOk("hooks"); ok {
		if _, ok := v.([]interface{})[0].(map[string]interface{}); ok {
			hooks := v.([]interface{})[0].(map[string]interface{})

			options.Hooks = &scalr.HooksOptions{
				PreInit:   scalr.String(hooks["pre_init"].(string)),
				PrePlan:   scalr.String(hooks["pre_plan"].(string)),
				PostPlan:  scalr.String(hooks["post_plan"].(string)),
				PreApply:  scalr.String(hooks["pre_apply"].(string)),
				PostApply: scalr.String(hooks["post_apply"].(string)),
			}
		}
	}

	if v, ok := d.Get("var_files").([]interface{}); ok {
		varFiles := make([]string, 0)
		for _, varFile := range v {
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
				createLinkOption.Alias = scalr.String(alias.(string))
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
	d.Set("name", workspace.Name)
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("force_latest_run", workspace.ForceLatestRun)
	d.Set("operations", workspace.Operations)
	d.Set("execution_mode", workspace.ExecutionMode)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("environment_id", workspace.Environment.ID)
	d.Set("has_resources", workspace.HasResources)
	d.Set("auto_queue_runs", workspace.AutoQueueRuns)
	d.Set("var_files", workspace.VarFiles)

	if workspace.RunOperationTimeout != nil {
		d.Set("run_operation_timeout", &workspace.RunOperationTimeout)
	}

	if workspace.VcsProvider != nil {
		d.Set("vcs_provider_id", workspace.VcsProvider.ID)
	}

	if workspace.AgentPool != nil {
		d.Set("agent_pool_id", workspace.AgentPool.ID)
	}

	var mv string
	if workspace.ModuleVersion != nil {
		mv = workspace.ModuleVersion.ID
	}
	d.Set("module_version_id", mv)

	var createdBy []interface{}
	if workspace.CreatedBy != nil {
		createdBy = append(createdBy, map[string]interface{}{
			"username":  workspace.CreatedBy.Username,
			"email":     workspace.CreatedBy.Email,
			"full_name": workspace.CreatedBy.FullName,
		})
	}
	d.Set("created_by", createdBy)

	var vcsRepo []interface{}
	if workspace.VCSRepo != nil {
		vcsRepo = append(vcsRepo, map[string]interface{}{
			"branch":             workspace.VCSRepo.Branch,
			"identifier":         workspace.VCSRepo.Identifier,
			"path":               workspace.VCSRepo.Path,
			"trigger_prefixes":   workspace.VCSRepo.TriggerPrefixes,
			"dry_runs_enabled":   workspace.VCSRepo.DryRunsEnabled,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
		})
	}
	d.Set("vcs_repo", vcsRepo)

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
	d.Set("hooks", hooks)

	providerConfigurationLinks, err := getProviderConfigurationWorkspaceLinks(scalrClient, id)
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
	d.Set("provider_configuration", providerConfigurations)

	var tagIDs []string
	if len(workspace.Tags) != 0 {
		for _, tag := range workspace.Tags {
			tagIDs = append(tagIDs, tag.ID)
		}
	}
	d.Set("tag_ids", tagIDs)

	return nil
}

func resourceScalrWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("auto_apply") || d.HasChange("auto_queue_runs") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") || d.HasChange("force_latest_run") ||
		d.HasChange("vcs_repo") || d.HasChange("operations") || d.HasChange("execution_mode") ||
		d.HasChange("vcs_provider_id") || d.HasChange("agent_pool_id") ||
		d.HasChange("hooks") || d.HasChange("module_version_id") || d.HasChange("var_files") ||
		d.HasChange("run_operation_timeout") {
		// Create a new options struct.
		options := scalr.WorkspaceUpdateOptions{
			Name:           scalr.String(d.Get("name").(string)),
			AutoApply:      scalr.Bool(d.Get("auto_apply").(bool)),
			ForceLatestRun: scalr.Bool(d.Get("force_latest_run").(bool)),
			Hooks: &scalr.HooksOptions{
				PreInit:   scalr.String(""),
				PrePlan:   scalr.String(""),
				PostPlan:  scalr.String(""),
				PreApply:  scalr.String(""),
				PostApply: scalr.String(""),
			},
		}

		// Process all configured options.
		if operations, ok := d.GetOk("operations"); ok {
			options.Operations = scalr.Bool(operations.(bool))
		}

		if executionMode, ok := d.GetOk("execution_mode"); ok {
			options.ExecutionMode = scalr.WorkspaceExecutionModePtr(
				scalr.WorkspaceExecutionMode(executionMode.(string)),
			)
		}

		if autoQueueRunsI, ok := d.GetOk("auto_queue_runs"); ok {
			options.AutoQueueRuns = scalr.AutoQueueRunsModePtr(
				scalr.WorkspaceAutoQueueRuns(autoQueueRunsI.(string)),
			)
		}

		if tfVersion, ok := d.GetOk("terraform_version"); ok {
			options.TerraformVersion = scalr.String(tfVersion.(string))
		}

		if v, ok := d.Get("var_files").([]interface{}); ok {
			varFiles := make([]string, 0)
			for _, varFile := range v {
				varFiles = append(varFiles, varFile.(string))
			}
			options.VarFiles = varFiles
		}

		options.WorkingDirectory = scalr.String(d.Get("working_directory").(string))

		if runOperationTimeout, ok := d.GetOk("run_operation_timeout"); ok {
			options.RunOperationTimeout = scalr.Int(runOperationTimeout.(int))
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
				Identifier:        scalr.String(vcsRepo["identifier"].(string)),
				Branch:            scalr.String(vcsRepo["branch"].(string)),
				Path:              scalr.String(vcsRepo["path"].(string)),
				TriggerPrefixes:   &triggerPrefixes,
				DryRunsEnabled:    scalr.Bool(vcsRepo["dry_runs_enabled"].(bool)),
				IngressSubmodules: scalr.Bool(vcsRepo["ingress_submodules"].(bool)),
			}
		}

		// Get and assert the hooks
		if v, ok := d.GetOk("hooks"); ok {
			if _, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				hooks := v.([]interface{})[0].(map[string]interface{})

				options.Hooks = &scalr.HooksOptions{
					PreInit:   scalr.String(hooks["pre_init"].(string)),
					PrePlan:   scalr.String(hooks["pre_plan"].(string)),
					PostPlan:  scalr.String(hooks["post_plan"].(string)),
					PreApply:  scalr.String(hooks["pre_apply"].(string)),
					PostApply: scalr.String(hooks["post_apply"].(string)),
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
					linkCreateOption.Alias = scalr.String(v.(string))
					mapID = mapID + v.(string)
				}
				expectedLinks[mapID] = linkCreateOption

			}
		}

		currentLinks, err := getProviderConfigurationWorkspaceLinks(scalrClient, id)
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
	scalrClient *scalr.Client, workspaceId string,
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
