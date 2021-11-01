package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrWorkspaceCreate,
		Read:   resourceScalrWorkspaceRead,
		Update: resourceScalrWorkspaceUpdate,
		Delete: resourceScalrWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 3,
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

			"operations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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

			"vcs_repo": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				ConflictsWith: []string{"module_version_id"},
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

func resourceScalrWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the name, environment_id and vcs_provider_id.
	name := d.Get("name").(string)
	environmentID := d.Get("environment_id").(string)

	// Create a new options struct.
	options := scalr.WorkspaceCreateOptions{
		Name:        scalr.String(name),
		AutoApply:   scalr.Bool(d.Get("auto_apply").(bool)),
		Operations:  scalr.Bool(d.Get("operations").(bool)),
		Environment: &scalr.Environment{ID: environmentID},
		Hooks:       &scalr.HooksOptions{},
	}

	// Process all configured options.
	if tfVersion, ok := d.GetOk("terraform_version"); ok {
		options.TerraformVersion = scalr.String(tfVersion.(string))
	}

	if workingDir, ok := d.GetOk("working_directory"); ok {
		options.WorkingDirectory = scalr.String(workingDir.(string))
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
			return err
		}

		options.VCSRepo = &scalr.VCSRepoOptions{
			Identifier:      scalr.String(vcsRepo["identifier"].(string)),
			Path:            scalr.String(vcsRepo["path"].(string)),
			TriggerPrefixes: &triggerPrefixes,
			DryRunsEnabled:  scalr.Bool(vcsRepo["dry_runs_enabled"].(bool)),
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
				PrePlan:   scalr.String(hooks["pre_plan"].(string)),
				PostPlan:  scalr.String(hooks["post_plan"].(string)),
				PreApply:  scalr.String(hooks["pre_apply"].(string)),
				PostApply: scalr.String(hooks["post_apply"].(string)),
			}
		}
	}

	log.Printf("[DEBUG] Create workspace %s for environment: %s", name, environmentID)
	workspace, err := scalrClient.Workspaces.Create(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating workspace %s for environment %s: %v", name, environmentID, err)
	}
	d.SetId(workspace.ID)
	return resourceScalrWorkspaceRead(d, meta)
}

func resourceScalrWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	log.Printf("[DEBUG] Read configuration of workspace: %s", id)
	workspace, err := scalrClient.Workspaces.ReadByID(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			log.Printf("[DEBUG] Workspace %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of workspace %s: %v", id, err)
	}

	// Update the config.
	d.Set("name", workspace.Name)
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("operations", workspace.Operations)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("environment_id", workspace.Environment.ID)
	d.Set("has_resources", workspace.HasResources)

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
			"branch":           workspace.VCSRepo.Branch,
			"identifier":       workspace.VCSRepo.Identifier,
			"path":             workspace.VCSRepo.Path,
			"trigger_prefixes": workspace.VCSRepo.TriggerPrefixes,
			"dry_runs_enabled": workspace.VCSRepo.DryRunsEnabled,
		})
	}
	d.Set("vcs_repo", vcsRepo)

	var hooks []interface{}
	if workspace.Hooks != nil {
		hooks = append(hooks, map[string]interface{}{
			"pre_plan":   workspace.Hooks.PrePlan,
			"post_plan":  workspace.Hooks.PostPlan,
			"pre_apply":  workspace.Hooks.PreApply,
			"post_apply": workspace.Hooks.PostApply,
		})
	}
	d.Set("hooks", hooks)

	return nil
}

func resourceScalrWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("auto_apply") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") ||
		d.HasChange("vcs_repo") || d.HasChange("operations") ||
		d.HasChange("vcs_provider_id") || d.HasChange("agent_pool_id") ||
		d.HasChange("hooks") || d.HasChange("module_version_id") {
		// Create a new options struct.
		options := scalr.WorkspaceUpdateOptions{
			Name:       scalr.String(d.Get("name").(string)),
			AutoApply:  scalr.Bool(d.Get("auto_apply").(bool)),
			Operations: scalr.Bool(d.Get("operations").(bool)),
			Hooks: &scalr.HooksOptions{
				PrePlan:   scalr.String(""),
				PostPlan:  scalr.String(""),
				PreApply:  scalr.String(""),
				PostApply: scalr.String(""),
			},
		}

		// Process all configured options.
		if tfVersion, ok := d.GetOk("terraform_version"); ok {
			options.TerraformVersion = scalr.String(tfVersion.(string))
		}

		options.WorkingDirectory = scalr.String(d.Get("working_directory").(string))

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
				return err
			}

			options.VCSRepo = &scalr.VCSRepoOptions{
				Identifier:      scalr.String(vcsRepo["identifier"].(string)),
				Branch:          scalr.String(vcsRepo["branch"].(string)),
				Path:            scalr.String(vcsRepo["path"].(string)),
				TriggerPrefixes: &triggerPrefixes,
				DryRunsEnabled:  scalr.Bool(vcsRepo["dry_runs_enabled"].(bool)),
			}
		}

		// Get and assert the hooks
		if v, ok := d.GetOk("hooks"); ok {
			if _, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				hooks := v.([]interface{})[0].(map[string]interface{})

				options.Hooks = &scalr.HooksOptions{
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
			return fmt.Errorf(
				"Error updating workspace %s: %v", id, err)
		}
	}

	return resourceScalrWorkspaceRead(d, meta)
}

func resourceScalrWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete workspace %s", id)
	err := scalrClient.Workspaces.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return nil
		}
		return fmt.Errorf("Error deleting workspace %s: %v", id, err)
	}

	return nil
}
