package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrWorkspaceResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

			"ssh_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
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

						"ingress_submodules": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

			"external_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceScalrWorkspaceStateUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if _, ok := rawState["external_id"]; !ok {
		// Due to migration drift, schema-versionV0 can already contain 'id' field,
		// so we can skip V0->V1 the migration.
		return rawState, nil
	}
	rawState["id"] = rawState["external_id"]
	delete(rawState, "external_id")
	delete(rawState, "ssh_key_id")
	return rawState, nil
}

func resourceScalrWorkspaceResourceV1() *schema.Resource {
	return &schema.Resource{
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

func resourceScalrWorkspaceStateUpgradeV1(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState["vcs_repo"] != nil {
		vcsRepos := rawState["vcs_repo"].([]interface{})
		if len(vcsRepos) == 0 {
			return rawState, nil
		}
		vcsRepo := vcsRepos[0].(map[string]interface{})
		rawState["vcs_provider_id"] = vcsRepo["oauth_token_id"]
		delete(vcsRepo, "oauth_token_id")
		rawState["vcs_repo"] = []interface{}{vcsRepo}
	}
	return rawState, nil
}

func resourceScalrWorkspaceResourceV2() *schema.Resource {
	return &schema.Resource{
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

func resourceScalrWorkspaceStateUpgradeV2(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	delete(rawState, "queue_all_runs")
	return rawState, nil
}

func resourceScalrWorkspaceResourceV3() *schema.Resource {
	return &schema.Resource{
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

			"var_files": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		},
	}
}

func resourceScalrWorkspaceStateUpgradeV3(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState["operations"].(bool) {
		rawState["execution_mode"] = scalr.WorkspaceExecutionModeRemote
	} else {
		rawState["execution_mode"] = scalr.WorkspaceExecutionModeLocal
	}
	return rawState, nil
}

func resourceScalrWorkspaceResourceV4() *schema.Resource {
	return &schema.Resource{
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

			"var_files": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

			"queue_all_runs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
		},
	}
}

func resourceScalrWorkspaceStateUpgradeV4(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	scalrClient := meta.(*scalr.Client)

	wsID := rawState["id"].(string)
	workspace, err := scalrClient.Workspaces.ReadByID(ctx, wsID)
	if err != nil {
		return nil, fmt.Errorf("Error reading workspace %s: %v", wsID, err)
	}

	rawState["queue_all_runs"] = workspace.QueueAllRuns
	return rawState, nil
}
