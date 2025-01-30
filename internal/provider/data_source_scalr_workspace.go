package provider

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a single workspace.",
		ReadContext: dataSourceScalrWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "ID of the workspace.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},

			"name": {
				Description:  "Name of the workspace.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"environment_id": {
				Description: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"vcs_provider_id": {
				Description: "The identifier of a VCS provider in the format `vcs-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"module_version_id": {
				Description: "The identifier of a module version in the format `modver-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"agent_pool_id": {
				Description: "The identifier of an agent pool in the format `apool-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"auto_apply": {
				Description: "Boolean indicates if `terraform apply` will be automatically run when `terraform plan` ends without error.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"force_latest_run": {
				Description: "Boolean indicates if latest new run will be automatically raised in priority.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"deletion_protection_enabled": {
				Description: "Boolean, indicates if the workspace has the protection from an accidental state lost. If enabled and the workspace has resource, the deletion will not be allowed.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"operations": {
				Description: "Boolean indicates if the workspace is being used for remote execution.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"execution_mode": {
				Description: "Execution mode of the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"terraform_version": {
				Description: "The version of Terraform used for this workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"terragrunt": {
				Description: "List of terragrunt configurations in a workspace if set.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Description: "The version of terragrunt.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"use_run_all": {
							Description: "Boolean indicates if terragrunt should run all commands.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"include_external_dependencies": {
							Description: "Boolean indicates if terragrunt should include external dependencies.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},

			"iac_platform": {
				Description: "The IaC platform used for this workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"type": {
				Description: "The type of the Scalr Workspace environment.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"working_directory": {
				Description: "A relative path that Terraform will execute within.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"has_resources": {
				Description: "The presence of active terraform resources in the current state version.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"auto_queue_runs": {
				Description: "Indicates if runs have to be queued automatically when a new configuration version is uploaded." +
					"\n\n  Supported values are `skip_first`, `always`, `never`:" +
					"\n\n  * `skip_first` - after the very first configuration version is uploaded into the workspace the run will not be triggered. But the following configurations will do. This is the default behavior." +
					"\n  * `always` - runs will be triggered automatically on every upload of the configuration version." +
					"\n  * `never` - configuration versions are uploaded into the workspace, but runs will not be triggered.",
				Type:     schema.TypeString,
				Computed: true,
			},

			"hooks": {
				Description: "List of custom hooks in a workspace.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pre_init": {
							Description: "Script or action configured to call before init phase.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"pre_plan": {
							Description: "Script or action configured to call before plan phase.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"post_plan": {
							Description: "Script or action configured to call after plan phase.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"pre_apply": {
							Description: "Script or action configured to call before apply phase.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"post_apply": {
							Description: "Script or action configured to call after apply phase.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"vcs_repo": {
				Description: "If a workspace is linked to a VCS repository this block shows the details, otherwise `{}`",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "The reference to the VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"path": {
							Description: "Path within the repo, if any.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"dry_runs_enabled": {
							Description: "Boolean indicates the VCS-driven dry runs should run when the pull request to the configuration versions branch is created.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"ingress_submodules": {
							Description: "Designates whether to clone git submodules of the VCS repository.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},

			"tag_ids": {
				Description: "List of tag IDs associated with the workspace.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
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
		},
	}
}

func dataSourceScalrWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	workspaceID := d.Get("id").(string)
	name := d.Get("name").(string)
	environmentID := d.Get("environment_id").(string)

	options := scalr.WorkspaceListOptions{
		Include: "created-by",
		Filter:  &scalr.WorkspaceFilter{Environment: ptr(environmentID)},
	}

	if workspaceID != "" {
		options.Filter.Id = ptr(workspaceID)
	}

	if name != "" {
		options.Filter.Name = ptr(name)
	}

	log.Printf("[DEBUG] Read configuration of workspace with ID '%s', name '%s', and environment_id '%s'", workspaceID, name, environmentID)

	workspaces, err := scalrClient.Workspaces.List(ctx, options)
	if err != nil {
		return diag.Errorf("error retrieving workspace: %v", err)
	}
	if len(workspaces.Items) > 1 {
		return diag.FromErr(errors.New("Your query returned more than one result. Please try a more specific search criteria."))
	}
	if len(workspaces.Items) == 0 {
		return diag.Errorf("Could not find workspace with ID '%s', name '%s' and environment_id '%s'", workspaceID, name, environmentID)
	}

	workspace := workspaces.Items[0]

	// Update the config.
	_ = d.Set("name", workspace.Name)
	_ = d.Set("auto_apply", workspace.AutoApply)
	_ = d.Set("force_latest_run", workspace.ForceLatestRun)
	_ = d.Set("deletion_protection_enabled", workspace.DeletionProtectionEnabled)
	_ = d.Set("operations", workspace.Operations)
	_ = d.Set("execution_mode", workspace.ExecutionMode)
	_ = d.Set("terraform_version", workspace.TerraformVersion)
	_ = d.Set("iac_platform", workspace.IaCPlatform)
	_ = d.Set("type", workspace.EnvironmentType)
	_ = d.Set("working_directory", workspace.WorkingDirectory)
	_ = d.Set("has_resources", workspace.HasResources)
	_ = d.Set("auto_queue_runs", workspace.AutoQueueRuns)

	if workspace.ModuleVersion != nil {
		_ = d.Set("module_version_id", workspace.ModuleVersion.ID)
	}

	if workspace.VcsProvider != nil {
		_ = d.Set("vcs_provider_id", workspace.VcsProvider.ID)
	}

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
		vcsConfig := map[string]interface{}{
			"identifier":         workspace.VCSRepo.Identifier,
			"path":               workspace.VCSRepo.Path,
			"dry_runs_enabled":   workspace.VCSRepo.DryRunsEnabled,
			"ingress_submodules": workspace.VCSRepo.IngressSubmodules,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}
	_ = d.Set("vcs_repo", vcsRepo)

	var terragrunt []interface{}
	if workspace.Terragrunt != nil {
		terragruntConfig := map[string]interface{}{
			"version":                       workspace.Terragrunt.Version,
			"use_run_all":                   workspace.Terragrunt.UseRunAll,
			"include_external_dependencies": workspace.Terragrunt.IncludeExternalDependencies,
		}
		terragrunt = append(terragrunt, terragruntConfig)
	}
	_ = d.Set("terragrunt", terragrunt)

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
	_ = d.Set("hooks", hooks)

	var tags []string
	if len(workspace.Tags) != 0 {
		for _, tag := range workspace.Tags {
			tags = append(tags, tag.ID)
		}
	}
	_ = d.Set("tag_ids", tags)

	d.SetId(workspace.ID)

	return nil
}
