package scalr

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrPolicyGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrPolicyGroupRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"opa_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vcs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"branch": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"commit": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sha": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"message": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"author": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vcs_provider_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"enforced_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"environments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workspaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceScalrPolicyGroupRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// required fields
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.PolicyGroupListOptions{
		Account: accountID,
		Name:    name,
		Include: "vcs-revision,policies,environments,workspaces",
	}
	log.Printf("[DEBUG] Read configuration of policy group: %s/%s", accountID, name)

	pgl, err := scalrClient.PolicyGroups.List(ctx, options)
	if err != nil {
		return fmt.Errorf("error retrieving policy group: %v", err)
	}

	if pgl.TotalCount == 0 {
		return fmt.Errorf("policy group %s/%s not found", accountID, name)
	}

	pg := pgl.Items[0]

	// Update the configuration.
	d.Set("status", pg.Status)
	d.Set("error_message", pg.ErrorMessage)
	d.Set("opa_version", pg.OpaVersion)

	if pg.VcsProvider != nil {
		d.Set("vcs_provider_id", pg.VcsProvider.ID)
	}

	if pg.VCSRepo != nil {
		log.Printf("[DEBUG] Read vcs revision attributes of policy group: %s", pg.ID)
		var vcsConfig []map[string]interface{}

		vcs := map[string]interface{}{
			"repository_id": pg.VCSRepo.Identifier,
			"branch":        pg.VCSRepo.Branch,
			"path":          pg.VCSRepo.Path,
			"commit":        []map[string]interface{}{},
		}

		if pg.VcsRevision != nil {
			vcs["commit"] = []map[string]interface{}{
				{
					"sha":     pg.VcsRevision.CommitSha,
					"message": pg.VcsRevision.CommitMessage,
					"author": map[string]interface{}{
						"username": pg.VcsRevision.SenderUsername,
					},
				},
			}
		}

		d.Set("vcs", append(vcsConfig, vcs))
	}

	if len(pg.Policies) != 0 {
		var policies []map[string]interface{}
		for _, policy := range pg.Policies {
			policies = append(policies, map[string]interface{}{
				"name":           policy.Name,
				"enabled":        policy.Enabled,
				"enforced_level": policy.EnforcementLevel,
			})
		}
		d.Set("policies", policies)
	}

	if len(pg.Environments) != 0 {
		var envs []string
		for _, env := range pg.Environments {
			envs = append(envs, env.ID)
		}
		d.Set("environments", envs)
	}

	if len(pg.Workspaces) != 0 {
		var wss []string
		for _, ws := range pg.Workspaces {
			wss = append(wss, ws.ID)
		}
		d.Set("workspaces", wss)
	}

	d.SetId(pg.ID)

	return nil
}
