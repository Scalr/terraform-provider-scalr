package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrWebhook() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrWebhookRead,

		Schema: map[string]*schema.Schema{

			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
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

func dataSourceScalrWebhookRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the ID
	webhookID := d.Get("id").(string)
	webhookName := d.Get("name").(string)

	if webhookID == "" && webhookName == "" {
		return fmt.Errorf("At least one argument 'id' or 'name' is required, but no definitions was found")
	}

	if webhookID != "" && webhookName != "" {
		return fmt.Errorf("Attributes 'name' and 'id' can not be set at the same time")
	}

	accountId := d.Get("account_id").(string)

	var webhook *scalr.Webhook
	var err error

	if webhookID != "" {
		log.Printf("[DEBUG] Read configuration of webhook: %s", webhookID)
		webhook, err = scalrClient.Webhooks.Read(ctx, webhookID)
	} else {
		log.Printf("[DEBUG] Read configuration of webhook: %s", webhookName)
		options := GetWebhookByNameOptions{
			Name: &webhookName,
		}
		if accountId != "" {
			options.Account = &accountId
		}
		webhook, err = GetWebhookByName(options, scalrClient)
	}

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return fmt.Errorf("Could not find webhook %s: %v", webhookID, err)
		}
		return fmt.Errorf("Error retrieving webhook: %v", err)
	}

	// Update the config.
	d.Set("name", webhook.Name)
	d.Set("enabled", webhook.Enabled)
	d.Set("last_triggered_at", webhook.LastTriggeredAt)

	events := []string{}
	if webhook.Events != nil {
		for _, event := range webhook.Events {
			events = append(events, event.ID)
		}
	}
	d.Set("events", events)

	if webhook.Workspace != nil {
		d.Set("workspace_id", webhook.Workspace.ID)
	}
	if webhook.Environment != nil {
		d.Set("environment_id", webhook.Environment.ID)
	}
	if webhook.Endpoint != nil {
		d.Set("endpoint_id", webhook.Endpoint.ID)
	}
	d.SetId(webhook.ID)

	return nil
}
