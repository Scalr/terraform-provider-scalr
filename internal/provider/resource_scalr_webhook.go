package provider

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func resourceScalrWebhook() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage the state of webhooks in Scalr. Creates, updates and destroy.",
		CreateContext: resourceScalrWebhookCreate,
		ReadContext:   resourceScalrWebhookRead,
		UpdateContext: resourceScalrWebhookUpdate,
		DeleteContext: resourceScalrWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourceScalrWebhookResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrWebhookStateUpgradeV0,
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "Name of the webhook.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},

			"enabled": {
				Description: "Set (true/false) to enable/disable the webhook.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},

			"last_triggered_at": {
				Description: "Date/time when webhook was last triggered.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"events": {
				Description: "List of event IDs.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice(
							[]string{"run:completed", "run:errored", "run:needs_attention"},
							false,
						),
					),
				},
				Required: true,
				MinItems: 1,
			},

			"url": {
				Description:      "Endpoint URL. Required if `endpoint_id` is not set.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},

			"secret_key": {
				Description: "Secret key to sign the webhook payload.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},

			"timeout": {
				Description:      "Endpoint timeout (in seconds).",
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          15,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 365*24*3600)),
			},

			"max_attempts": {
				Description:      "Max delivery attempts of the payload.",
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          3,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 1000)),
			},

			"header": {
				Description: "Additional headers to set in the webhook request.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description:  "The name of the header.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"value": {
							Description:  "The value of the header.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
					},
				},
			},

			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},

			"environments": {
				Description: "The list of environment identifiers that the webhook is shared to. Use `[\"*\"]` to share with all environments.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func parseEventDefinitions(d *schema.ResourceData) ([]*scalr.EventDefinition, error) {
	eventDefinitions := make([]*scalr.EventDefinition, 0)

	eventIds := d.Get("events").(*schema.Set).List()
	err := ValidateIDsDefinitions(eventIds)
	if err != nil {
		return nil, fmt.Errorf("Got error during parsing events: %s", err.Error())
	}

	for _, eventID := range eventIds {
		id := eventID.(string)
		eventDefinitions = append(eventDefinitions, &scalr.EventDefinition{ID: id})
	}

	return eventDefinitions, nil
}

func parseHeaders(d *schema.ResourceData) []*scalr.WebhookHeader {
	headers := d.Get("header").(*schema.Set)
	headerValues := make([]*scalr.WebhookHeader, 0)
	for _, headerI := range headers.List() {
		header := headerI.(map[string]interface{})
		headerValues = append(headerValues, &scalr.WebhookHeader{
			Name:  header["name"].(string),
			Value: header["value"].(string),
		})
	}
	return headerValues
}

func createWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {
	name := d.Get("name").(string)
	accountId := d.Get("account_id").(string)

	if accountId == "" {
		return fmt.Errorf("Attribute `account_id` is required when creating new-style webhook")
	}

	eventDefinitions, err := parseEventDefinitions(d)
	if err != nil {
		return err
	}

	options := scalr.WebhookIntegrationCreateOptions{
		Name:        &name,
		Url:         ptr(d.Get("url").(string)),
		Account:     &scalr.Account{ID: accountId},
		Events:      eventDefinitions,
		Enabled:     ptr(d.Get("enabled").(bool)),
		Timeout:     ptr(d.Get("timeout").(int)),
		MaxAttempts: ptr(d.Get("max_attempts").(int)),
	}

	if secretKey, ok := d.GetOk("secret_key"); ok {
		options.SecretKey = ptr(secretKey.(string))
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = ptr(true)
		} else if len(environments) > 0 {
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				if env.(string) == "*" {
					return fmt.Errorf(
						"You cannot simultaneously enable the webhook for all and a limited list of environments. Please remove either wildcard or environment identifiers.",
					)
				}
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			options.Environments = environmentValues
		}
	}

	if _, ok := d.GetOk("header"); ok {
		options.Headers = parseHeaders(d)
	}

	webhook, err := scalrClient.WebhookIntegrations.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating webhook %s: %v", name, err)
	}

	d.SetId(webhook.ID)
	// Secret key could be generated by the API and is returned only while creation.
	_ = d.Set("secret_key", webhook.SecretKey)

	return nil
}

func resourceScalrWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	err := createWebhook(ctx, d, scalrClient)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalrWebhookRead(ctx, d, meta)
}

func readWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {
	webhookID := d.Id()

	webhook, err := scalrClient.WebhookIntegrations.Read(ctx, webhookID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return fmt.Errorf("Could not find webhook %s: %v", webhookID, err)
		}
		return fmt.Errorf("Error retrieving webhook: %v", err)
	}

	// Update the config.
	_ = d.Set("name", webhook.Name)
	_ = d.Set("account_id", webhook.Account.ID)
	_ = d.Set("enabled", webhook.Enabled)
	_ = d.Set("last_triggered_at", webhook.LastTriggeredAt)
	_ = d.Set("url", webhook.Url)
	_ = d.Set("timeout", webhook.Timeout)
	_ = d.Set("max_attempts", webhook.MaxAttempts)

	events := make([]string, 0)
	if webhook.Events != nil {
		for _, event := range webhook.Events {
			events = append(events, event.ID)
		}
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

	return nil
}

func resourceScalrWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	if err := readWebhook(ctx, d, scalrClient); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {

	options := scalr.WebhookIntegrationUpdateOptions{}

	if d.HasChange("name") {
		options.Name = ptr(d.Get("name").(string))
	}

	if d.HasChange("url") {
		options.Url = ptr(d.Get("url").(string))
	}

	if d.HasChange("enabled") {
		options.Enabled = ptr(d.Get("enabled").(bool))
	}

	if d.HasChange("secret_key") {
		options.SecretKey = ptr(d.Get("secret_key").(string))
	}

	if d.HasChange("timeout") {
		options.Timeout = ptr(d.Get("timeout").(int))
	}

	if d.HasChange("max_attempts") {
		options.MaxAttempts = ptr(d.Get("max_attempts").(int))
	}

	if d.HasChange("header") {
		options.Headers = parseHeaders(d)
	}

	eventDefinitions, err := parseEventDefinitions(d)
	if err != nil {
		return err
	}
	options.Events = eventDefinitions

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = ptr(true)
			options.Environments = make([]*scalr.Environment, 0)
		} else {
			options.IsShared = ptr(false)
			environmentValues := make([]*scalr.Environment, 0)
			for _, env := range environments {
				if env.(string) == "*" {
					return fmt.Errorf(
						"You cannot simultaneously enable the webhook for all and a limited list of environments. Please remove either wildcard or environment identifiers.",
					)
				}
				environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
			}
			options.Environments = environmentValues
		}
	} else {
		options.IsShared = ptr(false)
		options.Environments = make([]*scalr.Environment, 0)
	}

	log.Printf("[DEBUG] Update webhook: %s", d.Id())
	_, err = scalrClient.WebhookIntegrations.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating webhook %s: %v", d.Id(), err)
	}

	return nil
}

func resourceScalrWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	err := updateWebhook(ctx, d, scalrClient)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalrWebhookRead(ctx, d, meta)
}

func resourceScalrWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete webhook: %s", d.Id())
	err := scalrClient.WebhookIntegrations.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting webhook %s: %v", d.Id(), err)
	}

	return nil
}
