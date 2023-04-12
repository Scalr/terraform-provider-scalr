package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrWebhookRead,

		Schema: map[string]*schema.Schema{

			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"last_triggered_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"endpoint_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceScalrWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get IDs
	webhookID := d.Get("id").(string)
	webhookName := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	var webhook *scalr.Webhook
	var err error

	log.Printf("[DEBUG] Read configuration of webhook with ID '%s' and name '%s'", webhookID, webhookName)
	if webhookID != "" {
		webhook, err = scalrClient.Webhooks.Read(ctx, webhookID)
		if err != nil {
			return diag.Errorf("Error retrieving webhook: %v", err)
		}
		if webhookName != "" && webhookName != webhook.Name {
			return diag.Errorf("Could not find webhook with ID '%s' and name '%s'", webhookID, webhookName)
		}
	} else {
		options := GetWebhookByNameOptions{
			Name:    &webhookName,
			Account: &accountID,
		}
		webhook, err = GetWebhookByName(ctx, options, scalrClient)
		if err != nil {
			return diag.Errorf("Error retrieving webhook: %v", err)
		}
		if webhookID != "" && webhookID != webhook.ID {
			return diag.Errorf("Could not find webhook with ID '%s' and name '%s'", webhookID, webhookName)
		}
	}

	// Update the config.
	_ = d.Set("name", webhook.Name)
	_ = d.Set("enabled", webhook.Enabled)
	_ = d.Set("last_triggered_at", webhook.LastTriggeredAt)

	events := make([]string, 0)
	if webhook.Events != nil {
		for _, event := range webhook.Events {
			events = append(events, event.ID)
		}
	}
	_ = d.Set("events", events)

	if webhook.Workspace != nil {
		_ = d.Set("workspace_id", webhook.Workspace.ID)
	}
	if webhook.Environment != nil {
		_ = d.Set("environment_id", webhook.Environment.ID)
	}
	if webhook.Endpoint != nil {
		_ = d.Set("endpoint_id", webhook.Endpoint.ID)
	}
	d.SetId(webhook.ID)

	return nil
}
