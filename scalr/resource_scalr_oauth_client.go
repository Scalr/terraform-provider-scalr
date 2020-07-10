package scalr

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	scalr "github.com/scalr/go-scalr"
)

func resourceTFEOAuthClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOAuthClientCreate,
		Read:   resourceTFEOAuthClientRead,
		Delete: resourceTFEOAuthClientDelete,

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"api_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"http_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"oauth_token": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},

			"private_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},

			"service_provider": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.ServiceProviderAzureDevOpsServer),
						string(scalr.ServiceProviderAzureDevOpsServices),
						string(scalr.ServiceProviderBitbucket),
						string(scalr.ServiceProviderGithub),
						string(scalr.ServiceProviderGithubEE),
						string(scalr.ServiceProviderGitlab),
						string(scalr.ServiceProviderGitlabCE),
						string(scalr.ServiceProviderGitlabEE),
					},
					false,
				),
			},

			"oauth_token_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTFEOAuthClientCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the organization and provider.
	organization := d.Get("organization").(string)
	privateKey := d.Get("private_key").(string)
	serviceProvider := scalr.ServiceProviderType(d.Get("service_provider").(string))

	if serviceProvider == scalr.ServiceProviderAzureDevOpsServer && privateKey == "" {
		return fmt.Errorf("private_key is required for service_provider %s", serviceProvider)
	}

	// Create a new options struct.
	options := scalr.OAuthClientCreateOptions{
		APIURL:          scalr.String(d.Get("api_url").(string)),
		HTTPURL:         scalr.String(d.Get("http_url").(string)),
		OAuthToken:      scalr.String(d.Get("oauth_token").(string)),
		PrivateKey:      scalr.String(privateKey),
		ServiceProvider: scalr.ServiceProvider(serviceProvider),
	}

	log.Printf("[DEBUG] Create an OAuth client for organization: %s", organization)
	oc, err := scalrClient.OAuthClients.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating OAuth client for organization %s: %v", organization, err)
	}

	d.SetId(oc.ID)

	return resourceTFEOAuthClientRead(d, meta)
}

func resourceTFEOAuthClientRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Read configuration of OAuth client: %s", d.Id())
	oc, err := scalrClient.OAuthClients.Read(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			log.Printf("[DEBUG] OAuth client %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	// Update the config.
	d.Set("api_url", oc.APIURL)
	d.Set("http_url", oc.HTTPURL)
	d.Set("organization", oc.Organization.Name)
	d.Set("service_provider", string(oc.ServiceProvider))

	switch len(oc.OAuthTokens) {
	case 0:
		d.Set("oauth_token_id", "")
	case 1:
		d.Set("oauth_token_id", oc.OAuthTokens[0].ID)
	default:
		return fmt.Errorf("Unexpected number of OAuth tokens: %d", len(oc.OAuthTokens))
	}

	return nil
}

func resourceTFEOAuthClientDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete OAuth client: %s", d.Id())
	err := scalrClient.OAuthClients.Delete(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting OAuth client %s: %v", d.Id(), err)
	}

	return nil
}
