package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrAgentPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrAgentPoolCreate,
		ReadContext:   resourceScalrAgentPoolRead,
		UpdateContext: resourceScalrAgentPoolUpdate,
		DeleteContext: resourceScalrAgentPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrAgentPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	var envID string

	// Get required options
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	// Create a new options struct
	options := scalr.AgentPoolCreateOptions{
		Name:    scalr.String(name),
		Account: &scalr.Account{ID: accountID},
	}

	if envID, ok := d.GetOk("environment_id"); ok {
		options.Environment = &scalr.Environment{
			ID: envID.(string),
		}
	}

	log.Printf("[DEBUG] Create agent pool %s for account: %s environment: %s", name, accountID, envID)
	agentPool, err := scalrClient.AgentPools.Create(ctx, options)
	if err != nil {
		return diag.Errorf(
			"Error creating agent pool %s for account %s environment %s: %v", name, accountID, envID, err)
	}
	d.SetId(agentPool.ID)
	return resourceScalrAgentPoolRead(ctx, d, meta)
}

func resourceScalrAgentPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()
	log.Printf("[DEBUG] Read configuration of agent pool: %s", id)
	agentPool, err := scalrClient.AgentPools.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] agent pool %s not found", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading configuration of agent pool %s: %v", id, err)
	}

	// Update the config.
	_ = d.Set("name", agentPool.Name)
	_ = d.Set("account_id", agentPool.Account.ID)

	if agentPool.Environment != nil {
		_ = d.Set("environment_id", agentPool.Environment.ID)
	} else {
		_ = d.Set("environment_id", nil)
	}
	return nil
}

func resourceScalrAgentPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	if d.HasChange("name") {
		// Create a new options struct
		options := scalr.AgentPoolUpdateOptions{
			Name: scalr.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Update agent pool %s", id)
		_, err := scalrClient.AgentPools.Update(ctx, id, options)
		if err != nil {
			return diag.Errorf(
				"Error updating agentPool %s: %v", id, err)
		}
	}

	return resourceScalrAgentPoolRead(ctx, d, meta)
}

func resourceScalrAgentPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete agent pool %s", id)
	err := scalrClient.AgentPools.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf(
			"Error deleting agent pool %s: %v", id, err)
	}

	return nil
}
