package provider

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func resourceScalrVcsProvider() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the Scalr VCS provider. Create, update and destroy.",
		CreateContext: resourceScalrVcsProviderCreate,
		ReadContext:   resourceScalrVcsProviderRead,
		UpdateContext: resourceScalrVcsProviderUpdate,
		DeleteContext: resourceVcsProviderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceScalrVcsProviderV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrVcsProviderStateUpgradeV0,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the vcs provider.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"url": {
				Description: "This field is required for self-hosted vcs providers.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"vcs_type": {
				Description: "The vcs provider type is one of `github`, `github_enterprise`, `gitlab`, `gitlab_enterprise`, `bitbucket_enterprise`. The other providers are not currently supported in the resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.Github),
						string(scalr.GithubEnterprise),
						string(scalr.Gitlab),
						string(scalr.GitlabEnterprise),
						string(scalr.BitbucketEnterprise),
					},
					false,
				),
			},
			"token": {
				Description: "The personal access token for the provider." +
					"\n  * GitHub token can be generated by url https://github.com/settings/tokens/new?description=example-vcs-resouce&scopes=repo" +
					"\n  * Gitlab token can be generated by url https://gitlab.com/-/profile/personal_access_tokens?name=example-vcs-resouce&scopes=api,read_user,read_registry",
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"username": {
				Description: "This field is required for `bitbucket_enterprise` provider type.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"account_id": {
				Description: "ID of the account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"agent_pool_id": {
				Description: "The id of the agent pool to connect Scalr to self-hosted VCS provider.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"environments": {
				Description: "The list of environment identifiers that the VCS provider is shared to. Use `[\"*\"]` to share with all environments.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"draft_pr_runs_enabled": {
				Description: "Enable draft PR runs for the VCS provider.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceScalrVcsProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	// Get attributes.
	name := d.Get("name").(string)
	token := d.Get("token").(string)
	vcsType := scalr.VcsType(d.Get("vcs_type").(string))
	options := scalr.VcsProviderCreateOptions{
		Name:     &name,
		VcsType:  vcsType,
		Token:    token,
		AuthType: "personal_token",
		Account:  &scalr.Account{ID: d.Get("account_id").(string)},
	}

	// Get the url
	if url, ok := d.GetOk("url"); ok {
		options.Url = ptr(url.(string))
	}

	// Get the username
	if username, ok := d.GetOk("username"); ok {
		options.Username = ptr(username.(string))
	}

	if agentPoolID, ok := d.GetOk("agent_pool_id"); ok {
		options.AgentPool = &scalr.AgentPool{
			ID: agentPoolID.(string),
		}
	}

	if draftPRsRunEnabled, ok := d.GetOk("draft_pr_runs_enabled"); ok {
		options.DraftPrRunsEnabled = ptr(draftPRsRunEnabled.(bool))
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = ptr(true)
		} else if len(environments) > 0 {
			options.IsShared = ptr(false)
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				if env.(string) == "*" {
					return diag.Errorf(
						"You cannot simultaneously enable the VCS provider for all and a limited list of environments. Please remove either wildcard or environment identifiers.",
					)
				}
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			options.Environments = environmentValues
		}
	}

	log.Printf("[DEBUG] Create vcs provider: %s", name)
	provider, err := scalrClient.VcsProviders.Create(ctx, options)
	if err != nil {
		return diag.Errorf("Error creating vcs provider %s: %v", name, err)
	}
	d.SetId(provider.ID)

	return resourceScalrVcsProviderRead(ctx, d, meta)
}

func resourceScalrVcsProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	providerID := d.Id()

	log.Printf("[DEBUG] Read vcs provider with ID: %s", providerID)
	provider, err := scalrClient.VcsProviders.Read(ctx, providerID)
	if err != nil {
		log.Printf("[DEBUG] vcs provider %s no longer exists", providerID)
		d.SetId("")
		return nil
	}
	_ = d.Set("name", provider.Name)
	_ = d.Set("url", provider.Url)
	_ = d.Set("vcs_type", provider.VcsType)
	_ = d.Set("username", provider.Username)
	if provider.Account != nil {
		_ = d.Set("account_id", provider.Account.ID)
	}
	if provider.AgentPool != nil {
		_ = d.Set("agent_pool_id", provider.AgentPool.ID)
	} else {
		_ = d.Set("agent_pool_id", "")
	}

	if provider.IsShared {
		allEnvironments := []string{"*"}
		_ = d.Set("environments", allEnvironments)
	} else {
		environmentIDs := make([]string, 0)
		for _, environment := range provider.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		_ = d.Set("environments", environmentIDs)
	}

	return nil
}

func resourceScalrVcsProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	// Create a new options' struct.
	options := scalr.VcsProviderUpdateOptions{
		Name:  ptr(d.Get("name").(string)),
		Token: ptr(d.Get("token").(string)),
	}

	if url, ok := d.GetOk("url"); ok {
		options.Url = ptr(url.(string))
	}

	// Get the username
	if username, ok := d.GetOk("username"); ok {
		options.Username = ptr(username.(string))
	}

	if agentPoolID, ok := d.GetOk("agent_pool_id"); ok {
		options.AgentPool = &scalr.AgentPool{
			ID: agentPoolID.(string),
		}
	}

	if d.HasChange("draft_pr_runs_enabled") {
		options.DraftPrRunsEnabled = ptr(d.Get("draft_pr_runs_enabled").(bool))
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = ptr(true)
			options.Environments = make([]*scalr.Environment, 0)
		} else {
			options.IsShared = ptr(false)
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				if env.(string) == "*" {
					return diag.Errorf(
						"You cannot simultaneously enable the VCS provider for all and a limited list of environments. Please remove either wildcard or environment identifiers.",
					)
				}
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			options.Environments = environmentValues
		}
	} else {
		options.IsShared = ptr(true)
		options.Environments = make([]*scalr.Environment, 0)
	}

	log.Printf("[DEBUG] Update vcs provider: %s", d.Id())
	_, err := scalrClient.VcsProviders.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf("Error updating vcs provider %s: %v", d.Id(), err)
	}

	return resourceScalrVcsProviderRead(ctx, d, meta)
}

func resourceVcsProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete vcs provider: %s", d.Id())
	err := scalrClient.VcsProviders.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting vcs provider %s: %v", d.Id(), err)
	}

	return nil
}
