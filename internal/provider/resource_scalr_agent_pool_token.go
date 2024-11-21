package provider

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrAgentPoolToken() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of agent pool's tokens in Scalr. Create, update and destroy.",
		CreateContext: resourceScalrAgentPoolTokenCreate,
		ReadContext:   resourceScalrAgentPoolTokenRead,
		UpdateContext: resourceScalrAgentPoolTokenUpdate,
		DeleteContext: resourceScalrAgentPoolTokenDelete,
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "Description of the token.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"agent_pool_id": {
				Description: "ID of the agent pool.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"token": {
				Description: "The token of the agent pool.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceScalrAgentPoolTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get required options
	poolID := d.Get("agent_pool_id").(string)

	// Create a new options struct
	options := scalr.AccessTokenCreateOptions{}

	if desc, ok := d.GetOk("description"); ok {
		options.Description = scalr.String(desc.(string))
	}

	log.Printf("[DEBUG] Create token for agent pool: %s", poolID)
	token, err := scalrClient.AgentPoolTokens.Create(ctx, poolID, options)
	if err != nil {
		return diag.Errorf(
			"Error creating token for agent pool %s: %v", poolID, err)
	}

	d.SetId(token.ID)
	// the token is returned from API only while creating
	_ = d.Set("token", token.Token)

	return resourceScalrAgentPoolTokenRead(ctx, d, meta)
}

func resourceScalrAgentPoolTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	poolID := d.Get("agent_pool_id").(string)

	if poolID == "" {
		return diag.Errorf("This resource does not support import")
	}

	log.Printf("[DEBUG] Read configuration of agent pool token: %s", id)
	options := scalr.AccessTokenListOptions{}

	for {
		tokensList, err := scalrClient.AgentPoolTokens.List(ctx, poolID, options)

		if err != nil {
			if errors.Is(err, scalr.ErrResourceNotFound) {
				log.Printf("[DEBUG] agent pool %s not found", poolID)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Error reading configuration of agent pool token %s: %v", id, err)
		}

		for _, t := range tokensList.Items {
			if t.ID == id {
				_ = d.Set("description", t.Description)
				return nil
			}
		}

		// Exit the loop when we've seen all pages.
		if tokensList.CurrentPage >= tokensList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = tokensList.NextPage
	}

	// the token has been deleted
	d.SetId("")
	return nil

}

func resourceScalrAgentPoolTokenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("description") {
		desc := d.Get("description").(string)
		// Create a new options struct
		options := scalr.AccessTokenUpdateOptions{
			Description: scalr.String(desc),
		}

		log.Printf("[DEBUG] Update agent pool token %s", id)
		_, err := scalrClient.AccessTokens.Update(ctx, id, options)
		if err != nil {
			return diag.Errorf(
				"Error updating agent pool token %s: %v", id, err)
		}
	}

	return resourceScalrAgentPoolTokenRead(ctx, d, meta)
}

func resourceScalrAgentPoolTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete agent pool token %s", id)
	err := scalrClient.AccessTokens.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting agent pool token %s: %v", id, err)
	}

	return nil
}
