package scalr

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrWebhookResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"last_triggered_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"events": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},

			"endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceScalrWebhookStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	scalrClient := meta.(*scalr.Client)
	webhookId := rawState["id"].(string)

	webhook, err := scalrClient.WebhookIntegrations.Read(ctx, webhookId)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of webhook %s: %w", webhookId, err)
	}

	rawState["account_id"] = webhook.Account.ID
	rawState["url"] = webhook.Url
	rawState["secret_key"] = webhook.SecretKey
	rawState["timeout"] = webhook.Timeout
	rawState["max_attempts"] = webhook.MaxAttempts

	headers := make([]map[string]interface{}, 0)
	if webhook.Headers != nil {
		for _, header := range webhook.Headers {
			headers = append(headers, map[string]interface{}{
				"name":  header.Name,
				"value": header.Value,
			})
		}
	}
	rawState["header"] = headers

	if webhook.IsShared {
		rawState["environments"] = []string{"*"}
	} else {
		environmentIDs := make([]string, len(webhook.Environments))
		for _, environment := range webhook.Environments {
			environmentIDs = append(environmentIDs, environment.ID)
		}
		rawState["environments"] = environmentIDs
	}

	return rawState, nil
}
