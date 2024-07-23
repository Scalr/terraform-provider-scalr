package scalr

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func resourceScalrEventBridgeIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of EventBridge integrations in Scalr. Create, update and destroy.",
		CreateContext: resourceScalrEventBridgeIntegrationCreate,
		ReadContext:   resourceScalrEventBridgeIntegrationRead,
		DeleteContext: resourceEventBridgeIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "Name of the EventBridge integration.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
			"aws_account_id": {
				Description:      "AWS account ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
			"region": {
				Description:      "AWS region.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
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

func resourceScalrEventBridgeIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	// Get attributes.
	name := d.Get("name").(string)
	awsAccountID := d.Get("aws_account_id").(string)
	region := d.Get("region").(string)

	options := scalr.EventBridgeIntegrationCreateOptions{
		Name:         &name,
		AWSAccountId: &awsAccountID,
		Region:       &region,
	}

	log.Printf("[DEBUG] Create EventBridge integration: %s", name)
	integration, err := scalrClient.EventBridgeIntegrations.Create(ctx, options)
	if err != nil {
		return diag.Errorf("Error creating EventBridge integration %s: %v", name, err)
	}
	d.SetId(integration.ID)

	return resourceScalrEventBridgeIntegrationRead(ctx, d, meta)
}

func resourceScalrEventBridgeIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	integrationID := d.Id()

	log.Printf("[DEBUG] Read EventBridge integration with ID: %s", integrationID)
	EventBridgeIntegration, err := scalrClient.EventBridgeIntegrations.Read(ctx, integrationID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] EventBridge integration %s no longer exists", integrationID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading EventBridge integration %s: %v", integrationID, err)
	}
	_ = d.Set("name", EventBridgeIntegration.Name)
	_ = d.Set("aws_account_id", EventBridgeIntegration.AWSAccountId)
	_ = d.Set("region", EventBridgeIntegration.Region)
	_ = d.Set("event_source_name", EventBridgeIntegration.EventSource)
	_ = d.Set("event_source_arn", EventBridgeIntegration.EventSourceARN)

	return nil
}

func resourceEventBridgeIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete EventBridge integration: %s", d.Id())
	err := scalrClient.EventBridgeIntegrations.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting EventBridge integration %s: %v", d.Id(), err)
	}

	return nil
}
