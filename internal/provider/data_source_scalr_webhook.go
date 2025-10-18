package provider

import (
	"context"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a webhook.",
		ReadContext: dataSourceScalrWebhookRead,

		Schema: map[string]*schema.Schema{

			"id": {
				Description:  "The webhook ID, in the format `wh-<RANDOM STRING>`.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},

			"name": {
				Description:  "Name of the webhook.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"enabled": {
				Description: "Boolean indicates if the webhook is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"last_triggered_at": {
				Description: "Date/time when webhook was last triggered.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"events": {
				Description: "List of event IDs.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},

			"url": {
				Description: "Endpoint URL.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"secret_key": {
				Description: "Secret key to sign the webhook payload.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Deprecated: "Attribute `secret_key` is deprecated, the secret-key has sensitive data" +
					" and is not returned by the API.",
			},

			"timeout": {
				Description: "Endpoint timeout (in seconds).",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"max_attempts": {
				Description: "Max delivery attempts of the payload.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"header": {
				Description: "Additional headers to set in the webhook request.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the header.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"value": {
							Description: "The value of the header.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},

			"environments": {
				Description: "The list of environment identifiers that the webhook is shared to, or `[\"*\"]` if shared with all environments.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
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

	var webhook *scalr.WebhookIntegration
	var err error

	log.Printf("[DEBUG] Read configuration of webhook with ID '%s' and name '%s'", webhookID, webhookName)
	// First read from new API by ID or search by name, as the new API
	// works both with old-style and new-style webhooks
	if webhookID != "" {
		webhook, err = scalrClient.WebhookIntegrations.Read(ctx, webhookID)
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
	_ = d.Set("account_id", webhook.Account.ID)
	_ = d.Set("enabled", webhook.Enabled)
	_ = d.Set("last_triggered_at", webhook.LastTriggeredAt)
	_ = d.Set("url", webhook.Url)
	_ = d.Set("secret_key", webhook.SecretKey)
	_ = d.Set("timeout", webhook.Timeout)
	_ = d.Set("max_attempts", webhook.MaxAttempts)

	events := make([]string, 0)
	if webhook.Events != nil {
		for _, event := range webhook.Events {
			events = append(events, event.ID)
		}
		sort.Strings(events)
	}
	_ = d.Set("events", events)

	headers := make([]map[string]interface{}, 0)
	if webhook.Headers != nil {
		for _, header := range webhook.Headers {
			headers = append(headers, map[string]interface{}{
				"name":  header.Name,
				"value": header.Value,
			})
		}
	}
	_ = d.Set("header", headers)

	if webhook.IsShared {
		_ = d.Set("environments", []string{"*"})
	} else {
		environmentIDs := make([]string, 0)
		for _, environment := range webhook.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		_ = d.Set("environments", environmentIDs)
	}

	d.SetId(webhook.ID)

	return nil
}
