package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrEventBridgeIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "This data source is used to retrieve details of a single EventBridge.",
		ReadContext: dataSourceScalrEventBridgeRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "ID of the EventBridge.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},
			"name": {
				Description:  "Name of the EventBridge.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"aws_account_id": {
				Description: "AWS account ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"region": {
				Description: "AWS region.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"event_source_name": {
				Description: "Event source name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"event_source_arn": {
				Description: "ARN of the event source.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceScalrEventBridgeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// required fields
	eventBridgeID := d.Get("id").(string)
	name := d.Get("name").(string)

	var eventBridge *scalr.EventBridgeIntegration
	var err error

	if eventBridgeID != "" {
		eventBridge, err = scalrClient.EventBridgeIntegrations.Read(ctx, eventBridgeID)
		if err != nil {
			return diag.Errorf("Error retrieving Event Bridge integration: %v", err)
		}
	} else {
		log.Printf("[DEBUG] Finding event bridge integration by name '%s'", name)
		options := scalr.EventBridgeIntegrationListOptions{
			Name: &name,
		}
		EventBridgeList, err := scalrClient.EventBridgeIntegrations.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving event bridge integrations: %v", err)
		}

		if len(EventBridgeList.Items) > 1 {
			return diag.FromErr(errors.New("Your query returned more than one result. Please try a more specific search criteria."))
		}

		if len(EventBridgeList.Items) == 0 {
			return diag.Errorf("Could not find event bridge integration with name '%s'", name)
		}

		eventBridge = EventBridgeList.Items[0]
	}

	// Update the config.
	_ = d.Set("name", eventBridge.Name)
	_ = d.Set("aws_account_id", eventBridge.AWSAccountId)
	_ = d.Set("region", eventBridge.Region)
	_ = d.Set("event_source_name", eventBridge.EventSource)
	_ = d.Set("event_source_arn", eventBridge.EventSourceARN)
	d.SetId(eventBridge.ID)

	return nil
}
