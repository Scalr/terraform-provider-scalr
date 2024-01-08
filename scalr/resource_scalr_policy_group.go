package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"log"
)

func resourceScalrPolicyGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of policy groups in Scalr. Create, update and destroy.",
		CreateContext: resourceScalrPolicyGroupCreate,
		ReadContext:   resourceScalrPolicyGroupRead,
		UpdateContext: resourceScalrPolicyGroupUpdate,
		DeleteContext: resourceScalrPolicyGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of a policy group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "A system status of the Policy group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"error_message": {
				Description: "A detailed error if Scalr failed to process the policy group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"opa_version": {
				Description: "The version of Open Policy Agent to run policies against. If omitted, the system default version is assigned.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"vcs_repo": {
				Description: "The VCS meta-data to create the policy from.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Description: "The reference to the VCS repository in the format `:org/:repo`, this refers to the organization and repository in your VCS provider.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"branch": {
							Description: "The branch of a repository the policy group is associated with. If omitted, the repository default branch will be used.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"path": {
							Description: "The subdirectory of the VCS repository where OPA policies are stored. If omitted or submitted as an empty string, this defaults to the repository's root.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"account_id": {
				Description: "The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"vcs_provider_id": {
				Description: "The identifier of a VCS provider, in the format `vcs-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"policies": {
				Description: "A list of the OPA policies the group verifies each run.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "A name of the policy.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"enabled": {
							Description: "If set to `false`, the policy will not be verified during a run.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"enforced_level": {
							Description: "An enforcement level of the policy.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"environments": {
				Description: "A list of the environments the policy group is linked to. Use `[\"*\"]` to enforce in all environments.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceScalrPolicyGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get required options
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)
	vcsProviderID := d.Get("vcs_provider_id").(string)
	vcsRepo := d.Get("vcs_repo").([]interface{})[0].(map[string]interface{})

	vcsOpt := &scalr.PolicyGroupVCSRepoOptions{
		Identifier: scalr.String(vcsRepo["identifier"].(string)),
	}
	if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
		vcsOpt.Branch = scalr.String(branch)
	}
	if path, ok := vcsRepo["path"].(string); ok && path != "" {
		vcsOpt.Path = scalr.String(path)
	}

	opts := scalr.PolicyGroupCreateOptions{
		Name:        scalr.String(name),
		VCSRepo:     vcsOpt,
		Account:     &scalr.Account{ID: accountID},
		VcsProvider: &scalr.VcsProvider{ID: vcsProviderID},
		IsEnforced:  scalr.Bool(false),
	}

	environments := make([]*scalr.Environment, 0)
	if environmentsI, ok := d.GetOk("environments"); ok {
		environmentsIDs := environmentsI.([]interface{})
		if (len(environmentsIDs) == 1) && environmentsIDs[0].(string) == "*" {
			opts.IsEnforced = scalr.Bool(true)
		} else if len(environmentsIDs) > 0 {
			for _, env := range environmentsIDs {
				if env.(string) == "*" {
					return diag.Errorf(
						"impossible to enforce the policy group in all and on a limited list of environments. Please remove either wildcard or environment identifiers",
					)
				}
				environments = append(environments, &scalr.Environment{ID: env.(string)})
			}
		}
	}

	// Optional attributes
	if opaVersion, ok := d.GetOk("opa_version"); ok {
		opts.OpaVersion = scalr.String(opaVersion.(string))
	}

	pg, err := scalrClient.PolicyGroups.Create(ctx, opts)
	if err != nil {
		return diag.Errorf("error creating policy group: %v", err)
	}

	d.SetId(pg.ID)

	if len(environments) > 0 && !*opts.IsEnforced {
		pgEnvs := make([]*scalr.PolicyGroupEnvironment, 0)
		for _, env := range environments {
			pgEnvs = append(pgEnvs, &scalr.PolicyGroupEnvironment{ID: env.ID})
		}
		pgEnvsOpts := scalr.PolicyGroupEnvironmentsCreateOptions{
			PolicyGroupID:           pg.ID,
			PolicyGroupEnvironments: pgEnvs,
		}
		err = scalrClient.PolicyGroupEnvironments.Create(ctx, pgEnvsOpts)
		if err != nil {
			defer func(ctx context.Context, pgID string) {
				_ = scalrClient.PolicyGroups.Delete(ctx, pgID)
			}(ctx, pg.ID)
			return diag.Errorf("error linking environments to policy group '%s': %v", name, err)
		}
	}

	return resourceScalrPolicyGroupRead(ctx, d, meta)
}

func resourceScalrPolicyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()
	log.Printf("[DEBUG] Read configuration of policy group %s", id)
	pg, err := scalrClient.PolicyGroups.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Policy group %s not found", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading configuration of policy group %s: %v", id, err)
	}

	// Update the configuration.
	_ = d.Set("name", pg.Name)
	_ = d.Set("status", pg.Status)
	_ = d.Set("error_message", pg.ErrorMessage)
	_ = d.Set("opa_version", pg.OpaVersion)
	_ = d.Set("account_id", pg.Account.ID)
	_ = d.Set("vcs_provider_id", pg.VcsProvider.ID)
	_ = d.Set("vcs_repo", []map[string]interface{}{{
		"identifier": pg.VCSRepo.Identifier,
		"branch":     pg.VCSRepo.Branch,
		"path":       pg.VCSRepo.Path,
	}})

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
		allEnvironments := []string{"*"}
		_ = d.Set("environments", allEnvironments)
	} else {
		environmentIDs := make([]string, 0)
		for _, environment := range pg.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		_ = d.Set("environments", environmentIDs)
	}

	return nil
}

func resourceScalrPolicyGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("opa_version") ||
		d.HasChange("vcs_provider_id") || d.HasChange("vcs_repo") ||
		d.HasChange("environments") {

		name := d.Get("name").(string)
		vcsProviderID := d.Get("vcs_provider_id").(string)
		vcsRepo := d.Get("vcs_repo").([]interface{})[0].(map[string]interface{})

		vcsOpt := &scalr.PolicyGroupVCSRepoOptions{
			Identifier: scalr.String(vcsRepo["identifier"].(string)),
		}
		if branch, ok := vcsRepo["branch"].(string); ok && branch != "" {
			vcsOpt.Branch = scalr.String(branch)
		}
		if path, ok := vcsRepo["path"].(string); ok && path != "" {
			vcsOpt.Path = scalr.String(path)
		}

		opts := scalr.PolicyGroupUpdateOptions{
			Name:        scalr.String(name),
			VCSRepo:     vcsOpt,
			VcsProvider: &scalr.VcsProvider{ID: vcsProviderID},
			IsEnforced:  scalr.Bool(false),
		}
		if opaVersion, ok := d.GetOk("opa_version"); ok {
			opts.OpaVersion = scalr.String(opaVersion.(string))
		}

		environments := make([]*scalr.Environment, 0)
		if environmentsI, ok := d.GetOk("environments"); ok {
			environmentsIDs := environmentsI.([]interface{})
			if (len(environmentsIDs) == 1) && environmentsIDs[0].(string) == "*" {
				opts.IsEnforced = scalr.Bool(true)
			} else if len(environmentsIDs) > 0 {
				for _, env := range environmentsIDs {
					if env.(string) == "*" {
						return diag.Errorf(
							"impossible to enforce the policy group in all and on a limited list of environments. Please remove either wildcard or environment identifiers",
						)
					}
					environments = append(environments, &scalr.Environment{ID: env.(string)})
				}
			}
		}

		log.Printf("[DEBUG] Update policy group %s", id)
		_, err := scalrClient.PolicyGroups.Update(ctx, id, opts)
		if err != nil {
			return diag.Errorf("error updating policy group %s: %v", id, err)
		}

		if !*opts.IsEnforced {
			pgEnvs := make([]*scalr.PolicyGroupEnvironment, 0)
			for _, env := range environments {
				pgEnvs = append(pgEnvs, &scalr.PolicyGroupEnvironment{ID: env.ID})
			}
			pgEnvsOpts := scalr.PolicyGroupEnvironmentsUpdateOptions{
				PolicyGroupID:           id,
				PolicyGroupEnvironments: pgEnvs,
			}

			err = scalrClient.PolicyGroupEnvironments.Update(ctx, pgEnvsOpts)
			if err != nil {
				return diag.Errorf("error updating environments for policy group %s: %v", id, err)
			}
		}
	}

	return resourceScalrPolicyGroupRead(ctx, d, meta)
}

func resourceScalrPolicyGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete policy group %s", id)
	err := scalrClient.PolicyGroups.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Policy group %s not found", id)
			return nil
		}
		return diag.Errorf("error deleting policy group %s: %v", id, err)
	}

	return nil
}
