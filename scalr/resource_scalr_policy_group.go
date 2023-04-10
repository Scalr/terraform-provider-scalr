package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrPolicyGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrPolicyGroupCreate,
		ReadContext:   resourceScalrPolicyGroupRead,
		UpdateContext: resourceScalrPolicyGroupUpdate,
		DeleteContext: resourceScalrPolicyGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
				Optional: true,
				Computed: true,
			},
			"vcs_repo": {
				Type:     schema.TypeList,
				Required: true,
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
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"vcs_provider_id": {
				Type:     schema.TypeString,
				Required: true,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
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

	var envs []string
	if len(pg.Environments) != 0 {
		for _, env := range pg.Environments {
			envs = append(envs, env.ID)
		}
	}
	_ = d.Set("environments", envs)

	return nil
}

func resourceScalrPolicyGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") || d.HasChange("opa_version") ||
		d.HasChange("vcs_provider_id") || d.HasChange("vcs_repo") {

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
		}
		if opaVersion, ok := d.GetOk("opa_version"); ok {
			opts.OpaVersion = scalr.String(opaVersion.(string))
		}

		log.Printf("[DEBUG] Update policy group %s", id)
		_, err := scalrClient.PolicyGroups.Update(ctx, id, opts)
		if err != nil {
			return diag.Errorf("error updating policy group %s: %v", id, err)
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
