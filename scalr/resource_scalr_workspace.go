package scalr

import (
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

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceScalrWorkspaceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrWorkspaceStateUpgradeV0,
				Version: 0,
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

			"queue_all_runs": {
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
				Computed: true,
			},

			"vcs_repo": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
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

						"oauth_token_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
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

func resourceScalrWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the name, environment_id and vcs-provider is.
	name := d.Get("name").(string)
	environmentID := d.Get("environment_id").(string)

	// Create a new options struct.
	options := scalr.WorkspaceCreateOptions{
		Name:         scalr.String(name),
		AutoApply:    scalr.Bool(d.Get("auto_apply").(bool)),
		Operations:   scalr.Bool(d.Get("operations").(bool)),
		QueueAllRuns: scalr.Bool(d.Get("queue_all_runs").(bool)),
	}

	// Process all configured options.
	if tfVersion, ok := d.GetOk("terraform_version"); ok {
		options.TerraformVersion = scalr.String(tfVersion.(string))
	}

	if workingDir, ok := d.GetOk("working_directory"); ok {
		options.WorkingDirectory = scalr.String(workingDir.(string))
	}

	if vcsProviderId, ok := d.GetOk("vcs_provider_id"); ok {
		options.VcsProvider = &scalr.VcsProviderOptions{
			ID: vcsProviderId.(string),
		}
	}

	// Get and assert the VCS repo configuration block.
	if v, ok := d.GetOk("vcs_repo"); ok {
		vcsRepo := v.([]interface{})[0].(map[string]interface{})

		options.VCSRepo = &scalr.VCSRepoOptions{
			Identifier:   scalr.String(vcsRepo["identifier"].(string)),
			Path:         scalr.String(vcsRepo["path"].(string)),
		}

		// Only set the branch if one is configured.
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			options.VCSRepo.Branch = scalr.String(branch)
		}
	}

	log.Printf("[DEBUG] Create workspace %s for environment: %s", name, environmentID)
	workspace, err := scalrClient.Workspaces.Create(ctx, environmentID, options)
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
		if err == scalr.ErrResourceNotFound {
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
	d.Set("queue_all_runs", workspace.QueueAllRuns)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("environment_id", workspace.Organization.Name)
	d.Set("vcs_provider_id", workspace.VcsProvider.ID)

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
		vcsConfig := map[string]interface{}{
			"identifier":     workspace.VCSRepo.Identifier,
			"path":           workspace.VCSRepo.Path,
		}

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			if vcsRepo, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				// Only set the branch if one is configured.
				if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
					vcsConfig["branch"] = workspace.VCSRepo.Branch
				}
			}
		}

		vcsRepo = append(vcsRepo, vcsConfig)
	}

	d.Set("vcs_repo", vcsRepo)
	return nil
}

func resourceScalrWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("auto_apply") || d.HasChange("queue_all_runs") ||
		d.HasChange("terraform_version") || d.HasChange("working_directory") || d.HasChange("vcs_repo") ||
		d.HasChange("operations") || d.HasChange("vcs_provider_id") {
		// Create a new options struct.
		options := scalr.WorkspaceUpdateOptions{
			Name:         scalr.String(d.Get("name").(string)),
			AutoApply:    scalr.Bool(d.Get("auto_apply").(bool)),
			Operations:   scalr.Bool(d.Get("operations").(bool)),
			QueueAllRuns: scalr.Bool(d.Get("queue_all_runs").(bool)),
		}

		// Process all configured options.
		if tfVersion, ok := d.GetOk("terraform_version"); ok {
			options.TerraformVersion = scalr.String(tfVersion.(string))
		}

		if workingDir, ok := d.GetOk("working_directory"); ok {
			options.WorkingDirectory = scalr.String(workingDir.(string))
		}

		if vcsProviderId, ok := d.GetOk("vcs_provider_id"); ok {
			options.VcsProvider = &scalr.VcsProviderOptions{
				ID: vcsProviderId.(string),
			}
		}

		// Get and assert the VCS repo configuration block.
		if v, ok := d.GetOk("vcs_repo"); ok {
			vcsRepo := v.([]interface{})[0].(map[string]interface{})

			options.VCSRepo = &scalr.VCSRepoOptions{
				Identifier:   scalr.String(vcsRepo["identifier"].(string)),
				Branch:       scalr.String(vcsRepo["branch"].(string)),
				Path:         scalr.String(vcsRepo["path"].(string)),
			}
		}

		log.Printf("[DEBUG] Update workspace %s", id)
		_, err := scalrClient.Workspaces.UpdateByID(ctx, id, options)
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
	err := scalrClient.Workspaces.DeleteByID(ctx, id)
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf(
			"Error deleting workspace %s: %v", id, err)
	}

	return nil
}
