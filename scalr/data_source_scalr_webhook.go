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

			"endpoint_id": {
				Description: "ID of the endpoint, in the format `ep-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated: "Attribute `endpoint_id` is deprecated, the endpoint information" +
					" is included in the `scalr_webhook` resource.",
			},

			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},

			"environment_id": {
				Description: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated: "The attribute `environment_id` is deprecated. The webhook is created on the" +
					" account level and the environments to which it is exposed" +
					" are controlled by the `environments` attribute.",
			},

			"workspace_id": {
				Description: "ID of the workspace, in the format `ws-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated:  "The attribute `workspace_id` is deprecated.",
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

	var newWebhook *scalr.WebhookIntegration
	var err error

	log.Printf("[DEBUG] Read configuration of webhook with ID '%s' and name '%s'", webhookID, webhookName)
	// First read from new API by ID or search by name, as the new API
	// works both with old-style and new-style webhooks
	if webhookID != "" {
		newWebhook, err = scalrClient.WebhookIntegrations.Read(ctx, webhookID)
		if err != nil {
			return diag.Errorf("Error retrieving webhook: %v", err)
		}
		if webhookName != "" && webhookName != newWebhook.Name {
			return diag.Errorf("Could not find webhook with ID '%s' and name '%s'", webhookID, webhookName)
		}
	} else {
		options := GetWebhookByNameOptions{
			Name:    &webhookName,
			Account: &accountID,
		}
		newWebhook, err = GetWebhookByName(ctx, options, scalrClient)
		if err != nil {
			return diag.Errorf("Error retrieving webhook: %v", err)
		}
		if webhookID != "" && webhookID != newWebhook.ID {
			return diag.Errorf("Could not find webhook with ID '%s' and name '%s'", webhookID, webhookName)
		}
	}
	// Having the webhook found, read from old API then
	// to populate deprecated fields available only in old API
	oldWebhook, err := scalrClient.Webhooks.Read(ctx, newWebhook.ID)
	if err != nil {
		return diag.Errorf("Error retrieving webhook: %v", err)
	}

	// Update the config.
	_ = d.Set("name", newWebhook.Name)
	_ = d.Set("account_id", newWebhook.Account.ID)
	_ = d.Set("enabled", newWebhook.Enabled)
	_ = d.Set("last_triggered_at", newWebhook.LastTriggeredAt)
	_ = d.Set("url", newWebhook.Url)
	_ = d.Set("secret_key", newWebhook.SecretKey)
	_ = d.Set("timeout", newWebhook.Timeout)
	_ = d.Set("max_attempts", newWebhook.MaxAttempts)

	events := make([]string, 0)
	if newWebhook.Events != nil {
		for _, event := range newWebhook.Events {
			events = append(events, event.ID)
		}
	}
	_ = d.Set("events", events)

	headers := make([]map[string]interface{}, 0)
	if newWebhook.Headers != nil {
		for _, header := range newWebhook.Headers {
			headers = append(headers, map[string]interface{}{
				"name":  header.Name,
				"value": header.Value,
			})
		}
	}
	_ = d.Set("header", headers)

	if newWebhook.IsShared {
		_ = d.Set("environments", []string{"*"})
	} else {
		environmentIDs := make([]string, 0)
		for _, environment := range newWebhook.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		_ = d.Set("environments", environmentIDs)
	}

	// Add deprecated attributes from old-style webhook
	if oldWebhook.Workspace != nil {
		_ = d.Set("workspace_id", oldWebhook.Workspace.ID)
	}
	if oldWebhook.Environment != nil {
		_ = d.Set("environment_id", oldWebhook.Environment.ID)
	}
	if oldWebhook.Endpoint != nil {
		_ = d.Set("endpoint_id", oldWebhook.Endpoint.ID)
	}

	d.SetId(newWebhook.ID)

	return nil
}
