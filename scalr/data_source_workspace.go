package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"vcs_provider_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"module_version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"agent_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_apply": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"operations": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"execution_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"terraform_version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"working_directory": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"has_resources": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"hooks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pre_init": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"pre_plan": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"post_plan": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"pre_apply": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"post_apply": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"vcs_repo": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dry_runs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"ingress_submodules": {
							Type:     schema.TypeBool,
							Computed: true,
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

func dataSourceScalrWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the name and environment_id.
	name := d.Get("name").(string)
	environmentID := d.Get("environment_id").(string)

	log.Printf("[DEBUG] Read configuration of workspace: %s", name)
	workspace, err := scalrClient.Workspaces.Read(ctx, environmentID, name)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return fmt.Errorf("Could not find workspace %s/%s", environmentID, name)
		}
		return fmt.Errorf("Error retrieving workspace: %v", err)
	}

	// Update the config.
	d.Set("auto_apply", workspace.AutoApply)
	d.Set("operations", workspace.Operations)
	d.Set("execution_mode", workspace.ExecutionMode)
	d.Set("terraform_version", workspace.TerraformVersion)
	d.Set("working_directory", workspace.WorkingDirectory)
	d.Set("has_resources", workspace.HasResources)

	if workspace.ModuleVersion != nil {
		d.Set("module_version_id", workspace.ModuleVersion.ID)
	}

	if workspace.VcsProvider != nil {
		d.Set("vcs_provider_id", workspace.VcsProvider.ID)
	}

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
			"identifier":         workspace.VCSRepo.Identifier,
			"path":               workspace.VCSRepo.Path,
			"dry_runs_enabled":   workspace.VCSRepo.DryRunsEnabled,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
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
	}
	d.Set("hooks", hooks)

	d.SetId(workspace.ID)

	return nil
}
