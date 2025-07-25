package provider

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrServiceAccountToken() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of service account's tokens in Scalr. Create, update and destroy.",
		CreateContext: resourceScalrServiceAccountTokenCreate,
		ReadContext:   resourceScalrServiceAccountTokenRead,
		UpdateContext: resourceScalrServiceAccountTokenUpdate,
		DeleteContext: resourceScalrServiceAccountTokenDelete,
		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Description: "ID of the service account.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Description of the token.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"token": {
				Description: "The token of the service account.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"name": {
				Description: "Name of the token.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"expires_in": {
				Description: "Number of minutes until the token expires.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceScalrServiceAccountTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	saID := d.Get("service_account_id").(string)

	options := scalr.AccessTokenCreateOptions{}
	if desc, ok := d.GetOk("description"); ok {
		options.Description = ptr(desc.(string))
	}
	if name, ok := d.GetOk("name"); ok {
		options.Name = ptr(name.(string))
	}
	if expiresIn, ok := d.GetOk("expires_in"); ok {
		options.ExpiresIn = ptr(expiresIn.(int))
	}

	log.Printf("[DEBUG] Create access token for service account: %s", saID)
	at, err := scalrClient.ServiceAccountTokens.Create(ctx, saID, options)
	if err != nil {
		return diag.Errorf(
			"Error creating access token for service account %s: %v", saID, err)
	}

	// the token is returned from API only while creating
	_ = d.Set("token", at.Token)

	d.SetId(at.ID)

	return resourceScalrServiceAccountTokenRead(ctx, d, meta)
}

func resourceScalrServiceAccountTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	saID := d.Get("service_account_id").(string)

	if saID == "" {
		return diag.Errorf("This resource does not support import")
	}

	log.Printf("[DEBUG] Read service account token: %s", id)
	options := scalr.AccessTokenListOptions{}

	for {
		atl, err := scalrClient.ServiceAccountTokens.List(ctx, saID, options)

		if err != nil {
			if errors.Is(err, scalr.ErrResourceNotFound) {
				log.Printf("[DEBUG] service account %s not found", saID)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Error reading service account token %s: %v", id, err)
		}

		for _, at := range atl.Items {
			if at.ID == id {
				_ = d.Set("description", at.Description)
				_ = d.Set("name", at.Name)
				_ = d.Set("expires_in", at.ExpiresIn)
				return nil
			}
		}

		// Exit the loop when we've seen all pages.
		if atl.CurrentPage >= atl.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = atl.NextPage
	}

	// the token has been deleted
	d.SetId("")
	return nil
}

func resourceScalrServiceAccountTokenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("description") || d.HasChange("name") {
		desc := d.Get("description").(string)
		name := d.Get("name").(string)

		options := scalr.AccessTokenUpdateOptions{
			Description: ptr(desc),
			Name:        ptr(name),
		}

		log.Printf("[DEBUG] Update service account access token %s", id)
		_, err := scalrClient.AccessTokens.Update(ctx, id, options)
		if err != nil {
			return diag.Errorf(
				"Error updating service account access token %s: %v", id, err)
		}
	}

	return resourceScalrServiceAccountTokenRead(ctx, d, meta)
}

func resourceScalrServiceAccountTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete service account access token %s", id)
	err := scalrClient.AccessTokens.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting service account access token %s: %v", id, err)
	}

	return nil
}
