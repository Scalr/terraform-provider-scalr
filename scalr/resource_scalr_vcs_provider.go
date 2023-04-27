package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func resourceScalrVcsProvider() *schema.Resource {
	return &schema.Resource{
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
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vcs_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"agent_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environments": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		options.Url = scalr.String(url.(string))
	}

	// Get the username
	if username, ok := d.GetOk("username"); ok {
		options.Username = scalr.String(username.(string))
	}

	if agentPoolID, ok := d.GetOk("agent_pool_id"); ok {
		options.AgentPool = &scalr.AgentPool{
			ID: agentPoolID.(string),
		}
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = true
		} else if len(environments) > 0 {
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
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
		Name:  scalr.String(d.Get("name").(string)),
		Token: scalr.String(d.Get("token").(string)),
	}

	if url, ok := d.GetOk("url"); ok {
		options.Url = scalr.String(url.(string))
	}

	// Get the username
	if username, ok := d.GetOk("username"); ok {
		options.Username = scalr.String(username.(string))
	}

	if agentPoolID, ok := d.GetOk("agent_pool_id"); ok {
		options.AgentPool = &scalr.AgentPool{
			ID: agentPoolID.(string),
		}
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = true
			options.Environments = make([]*scalr.Environment, 0)
		} else {
			options.IsShared = false
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			options.Environments = environmentValues
		}
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
