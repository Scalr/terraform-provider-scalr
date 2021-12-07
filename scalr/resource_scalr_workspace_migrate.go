package scalr

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
		// Due to migration drift, schema-versionV0 can already contain 'id' field
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

func resourceScalrWorkspaceStateUpgradeV3(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState["vcs_repo"] != nil {
		vcsRepos := rawState["vcs_repo"].([]interface{})
		if len(vcsRepos) == 0 {
			return rawState, nil
		}
		vcsRepo := vcsRepos[0].(map[string]interface{})
		path := vcsRepo["path"].(string)
		delete(vcsRepo, "path")

		if path != "" && path != "." {
			workingDir := rawState["working_directory"].(string)
			if workingDir != "" {
				workingDir = path + "/" + workingDir
			} else {
				workingDir = path
			}
			triggerPrefixes := vcsRepo["trigger_prefixes"].([]interface{})
			if len(triggerPrefixes) == 0 {
				if path != workingDir {
					triggerPrefixes = append(triggerPrefixes, path, workingDir)
				} else {
					triggerPrefixes = append(triggerPrefixes, path)
				}
			} else {
				for k, tp := range triggerPrefixes {
					triggerPrefixes[k] = path + "/" + tp.(string)
				}
			}
			vcsRepo["trigger_prefixes"] = triggerPrefixes
			rawState["working_directory"] = workingDir
		}
		rawState["vcs_repo"] = []interface{}{vcsRepo}
	}

	return rawState, nil
}
