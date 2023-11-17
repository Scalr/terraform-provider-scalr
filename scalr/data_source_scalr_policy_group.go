package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrPolicyGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a policy group.",
		ReadContext: dataSourceScalrPolicyGroupRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The identifier of a policy group.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},
			"name": {
				Description:  "The name of a policy group.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"status": {
				Description: "A system status of the policy group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"error_message": {
				Description: "An error details if Scalr failed to process the policy group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"opa_version": {
				Description: "The version of the Open Policy Agent that the policy group is using.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vcs_repo": {
				Description: "Contains VCS-related meta-data for the policy group.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "A reference to the VCS repository in the format `:org/:repo`, it stands for the organization and repository.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"branch": {
							Description: "A branch of a repository the policy group is associated with.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"path": {
							Description: "A subdirectory of a VCS repository where OPA policies are stored.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"account_id": {
				Description: "The identifier of the Scalr account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"vcs_provider_id": {
				Description: "The VCS provider identifier for the repository where the policy group resides. In the format `vcs-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"policies": {
				Description: "A list of the OPA policies the policy group verifies each run.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "A name of a policy.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"enabled": {
							Description: "If set to `false`, the policy will not be verified on a run.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"enforced_level": {
							Description: "An enforcement level of a policy.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"environments": {
				Description: "A list of the environments the policy group is linked to.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrPolicyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// required fields
	pgID := d.Get("id").(string)
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.PolicyGroupListOptions{
		Account: accountID,
		Include: "policies",
	}

	if pgID != "" {
		options.PolicyGroup = pgID
	}

	if name != "" {
		options.Name = name
	}

	log.Printf("[DEBUG] Read configuration of policy group with ID '%s', name '%s' and account_id '%s'", pgID, name, accountID)

	pgl, err := scalrClient.PolicyGroups.List(ctx, options)
	if err != nil {
		return diag.Errorf("error retrieving policy group: %v", err)
	}

	if pgl.TotalCount == 0 {
		return diag.Errorf("policy group with ID '%s', name '%s' and account_id '%s' not found", pgID, name, accountID)
	}

	pg := pgl.Items[0]

	// Update the configuration.
	_ = d.Set("name", pg.Name)
	_ = d.Set("status", pg.Status)
	_ = d.Set("error_message", pg.ErrorMessage)
	_ = d.Set("opa_version", pg.OpaVersion)

	if pg.VcsProvider != nil {
		_ = d.Set("vcs_provider_id", pg.VcsProvider.ID)
	}

	var vcsRepo []interface{}
	if pg.VCSRepo != nil {
		vcsConfig := map[string]interface{}{
			"identifier": pg.VCSRepo.Identifier,
			"branch":     pg.VCSRepo.Branch,
			"path":       pg.VCSRepo.Path,
		}
		vcsRepo = append(vcsRepo, vcsConfig)
	}
	_ = d.Set("vcs_repo", vcsRepo)

	var policies []map[string]interface{}
	if len(pg.Policies) != 0 {
		for _, policy := range pg.Policies {
			policies = append(policies, map[string]interface{}{
				"name":           policy.Name,
				"enabled":        policy.Enabled,
				"enforced_level": policy.EnforcementLevel,
			})
		}
	}
	_ = d.Set("policies", policies)

	if pg.IsEnforced {
		_ = d.Set("environments", []string{"*"})
	} else {
		envs := make([]string, 0)
		for _, env := range pg.Environments {
			envs = append(envs, env.ID)
		}
		_ = d.Set("environments", envs)
	}

	d.SetId(pg.ID)

	return nil
}
