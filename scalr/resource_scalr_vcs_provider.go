package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/scalr/go-scalr"
)

func resourceScalrVcsProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrVcsProviderCreate,
		Read:   resourceScalrVcsProviderRead,
		Update: resourceScalrVcsProviderUpdate,
		Delete: resourceVcsProviderDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrVcsProviderCreate(d *schema.ResourceData, meta interface{}) error {
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
	}

	// Get the url
	if url, ok := d.GetOk("url"); ok {
		options.Url = scalr.String(url.(string))
	}

	// Get the username
	if username, ok := d.GetOk("username"); ok {
		options.Username = scalr.String(username.(string))
	}

	// Get the account
	if accountId, ok := d.GetOk("account_id"); ok {
		options.Account = &scalr.Account{
			ID: accountId.(string),
		}
	}

	log.Printf("[DEBUG] Create vcs provider: %s", name)
	provider, err := scalrClient.VcsProviders.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating vcs provider %s: %v", name, err)
	}
	d.SetId(provider.ID)

	return resourceScalrVcsProviderRead(d, meta)
}

func resourceScalrVcsProviderRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	providerID := d.Id()

	log.Printf("[DEBUG] Read vcs provider with ID: %s", providerID)
	provider, err := scalrClient.VcsProviders.Read(ctx, providerID)
	if err != nil {
		return fmt.Errorf("Error retrieving vcs provider: %v", err)
	}
	d.Set("name", provider.Name)
	d.Set("url", provider.Url)
	d.Set("vcs_type", provider.VcsType)
	d.Set("auth_type", provider.AuthType)
	d.Set("username", provider.Username)
	if provider.Account != nil {
		d.Set("account_id", provider.Account.ID)
	}

	return nil
}

func resourceScalrVcsProviderUpdate(d *schema.ResourceData, meta interface{}) error {
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

	log.Printf("[DEBUG] Update vcs provider: %s", d.Id())
	_, err := scalrClient.VcsProviders.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating vcs provider %s: %v", d.Id(), err)
	}

	return resourceScalrVcsProviderRead(d, meta)
}

func resourceVcsProviderDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete vcs provider: %s", d.Id())
	err := scalrClient.VcsProviders.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting vcs provider %s: %v", d.Id(), err)
	}

	return nil
}
