package scalr

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
		CreateContext: resourceScalrWebhookCreate,
		ReadContext:   resourceScalrWebhookRead,
		UpdateContext: resourceScalrWebhookUpdate,
		DeleteContext: resourceScalrWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: forceRecreateIf(),
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"last_triggered_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"events": {
				Type: schema.TypeList,
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

			"endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
				Deprecated: "Attribute `endpoint_id` is deprecated, please set the endpoint information" +
					" in the webhook itself.",
				// If `endpoint_id` is set in configuration, we consider this an old-style webhook,
				// therefore using old API to create it.
				// That's why it conflicts with the fields that are only in new-style webhooks, so
				// user has two distinct sets of arguments for old and new webhooks.
				// One of `endpoint_id` or `url` must be set, and this defines which style will be chosen.
				ConflictsWith: []string{
					"url", "secret_key", "timeout", "max_attempts", "header", "environments", "account_id",
				},
				AtLeastOneOf: []string{"url"},
			},

			"url": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
				ConflictsWith:    []string{"endpoint_id", "workspace_id", "environment_id"},
			},

			"secret_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},

			"timeout": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          15,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 365*24*3600)),
			},

			"max_attempts": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          3,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 1000)),
			},

			"header": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
					},
				},
			},

			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: webhookAccountIDDefaultFunc,
				ForceNew:    true,
			},

			"workspace_id": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "The attribute `workspace_id` is deprecated.",
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Deprecated: "The attribute `environment_id` is deprecated. The webhook is created on the" +
					" account level and the environments to which it is exposed" +
					" are controlled by the `environments` attribute.",
				ConflictsWith: []string{"environments"},
			},

			"environments": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func forceRecreateIf() schema.CustomizeDiffFunc {
	// Destroy and recreate a webhook when `endpoint_id` has changed from having a value to unset,
	// which means switching from old-style to new-style webhook - and vice versa,
	// so we don't mix both style.
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		oldId, newId := d.GetChange("endpoint_id")
		if (oldId.(string) == "") != (newId.(string) == "") {
			return d.ForceNew("endpoint_id")
		}
		return nil
	}
}

// webhookAccountIDDefaultFunc returns default account id, if present.
// It won't return an error, as `account_id` is optional and computed
// for old-style webhooks.
func webhookAccountIDDefaultFunc() (interface{}, error) {
	accID, _ := getDefaultScalrAccountID()
	return accID, nil
}

// remove after https://scalr-labs.atlassian.net/browse/SCALRCORE-16234
func getResourceScope(ctx context.Context, scalrClient *scalr.Client, workspaceID string, environmentID string) (*scalr.Workspace, *scalr.Environment, *scalr.Account, error) {

	// Resource scope
	var workspace *scalr.Workspace
	var environment *scalr.Environment
	var account *scalr.Account

	// Get the workspace.
	if workspaceID != "" {
		var err error
		workspace, err = scalrClient.Workspaces.ReadByID(ctx, workspaceID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error retrieving workspace %s: %v", workspaceID, err)
		}

		if environmentID != "" && environmentID != workspace.Environment.ID {
			return nil, nil, nil, fmt.Errorf("Workspace %s does not belong to an environment %s", workspaceID, environmentID)
		}

		environmentID = workspace.Environment.ID
	}

	// Get the environment.
	if environmentID != "" {
		var err error
		environment, err = scalrClient.Environments.Read(ctx, environmentID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error retrieving environment %s: %v", environmentID, err)
		}
		account = environment.Account
	} else {
		return nil, nil, nil, fmt.Errorf("Missing workspace_id or environment_id")
	}

	return workspace, environment, account, nil
}

func parseEventDefinitions(d *schema.ResourceData) ([]*scalr.EventDefinition, error) {
	eventDefinitions := make([]*scalr.EventDefinition, 0)

	eventIds := d.Get("events").([]interface{})
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

func createOldWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {
	// Get attributes.
	name := d.Get("name").(string)
	endpointID := d.Get("endpoint_id").(string)
	workspaceID := d.Get("workspace_id").(string)
	environmentID := d.Get("environment_id").(string)

	workspace, environment, account, err := getResourceScope(ctx, scalrClient, workspaceID, environmentID)
	if err != nil {
		return err
	}

	eventDefinitions, err := parseEventDefinitions(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := scalr.WebhookCreateOptions{
		Name:        scalr.String(name),
		Enabled:     scalr.Bool(d.Get("enabled").(bool)),
		Events:      eventDefinitions,
		Endpoint:    &scalr.Endpoint{ID: endpointID},
		Workspace:   workspace,
		Environment: environment,
		Account:     account,
	}

	if workspaceID != "" {
		options.Workspace = &scalr.Workspace{ID: workspaceID}
	}
	if environmentID != "" {
		options.Environment = &scalr.Environment{ID: environmentID}
	}
	if environmentID != "" {
		options.Environment = &scalr.Environment{ID: environmentID}
	}

	log.Printf("[DEBUG] Create webhook: %s", name)
	webhook, err := scalrClient.Webhooks.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating webhook %s: %v", name, err)
	}

	d.SetId(webhook.ID)
	return nil
}

func createNewWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {
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
		Url:         scalr.String(d.Get("url").(string)),
		Account:     &scalr.Account{ID: accountId},
		Events:      eventDefinitions,
		Enabled:     scalr.Bool(d.Get("enabled").(bool)),
		Timeout:     scalr.Int(d.Get("timeout").(int)),
		MaxAttempts: scalr.Int(d.Get("max_attempts").(int)),
	}

	if secretKey, ok := d.GetOk("secret_key"); ok {
		options.SecretKey = scalr.String(secretKey.(string))
	}

	if environmentsI, ok := d.GetOk("environments"); ok {
		environments := environmentsI.(*schema.Set).List()
		if (len(environments) == 1) && (environments[0].(string) == "*") {
			options.IsShared = scalr.Bool(true)
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

	return nil
}

func resourceScalrWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	endpointID := d.Get("endpoint_id").(string)

	var err error
	// Here the old method is kept and is used to create old-style webhooks (with `endpoint_id` attribute set).
	// After deprecation period it should be easy to remove it completely.
	// Same for updating a webhook.
	if endpointID != "" {
		err = createOldWebhook(ctx, d, scalrClient)
	} else {
		err = createNewWebhook(ctx, d, scalrClient)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalrWebhookRead(ctx, d, meta)
}

func readOldWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {
	webhookID := d.Id()

	webhook, err := scalrClient.Webhooks.Read(ctx, webhookID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return fmt.Errorf("Could not find webhook %s: %v", webhookID, err)
		}
		return fmt.Errorf("Error retrieving webhook: %v", err)
	}

	if webhook.Workspace != nil {
		_ = d.Set("workspace_id", webhook.Workspace.ID)
	} else {
		_ = d.Set("workspace_id", nil)
	}
	if webhook.Environment != nil {
		_ = d.Set("environment_id", webhook.Environment.ID)
	} else {
		_ = d.Set("environment_id", nil)
	}
	if webhook.Endpoint != nil {
		_ = d.Set("endpoint_id", webhook.Endpoint.ID)
	} else {
		_ = d.Set("endpoint_id", nil)
	}

	return nil
}

func readNewWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {
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
	_ = d.Set("secret_key", webhook.SecretKey)
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

	// Reading the webhook differs from create or update methods.
	// We read from both old and new API and basically merge the fields from both resources,
	// therefore keeping deprecated attributes in place for now and extending with the new ones.
	if err := readOldWebhook(ctx, d, scalrClient); err != nil {
		return diag.FromErr(err)
	}
	if err := readNewWebhook(ctx, d, scalrClient); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateOldWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {
	eventDefinitions, err := parseEventDefinitions(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := scalr.WebhookUpdateOptions{
		Name:     scalr.String(d.Get("name").(string)),
		Enabled:  scalr.Bool(d.Get("enabled").(bool)),
		Events:   eventDefinitions,
		Endpoint: &scalr.Endpoint{ID: d.Get("endpoint_id").(string)},
	}

	log.Printf("[DEBUG] Update webhook: %s", d.Id())
	_, err = scalrClient.Webhooks.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating webhook %s: %v", d.Id(), err)
	}

	return nil
}

func updateNewWebhook(ctx context.Context, d *schema.ResourceData, scalrClient *scalr.Client) error {

	options := scalr.WebhookIntegrationUpdateOptions{}

	if d.HasChange("name") {
		options.Name = scalr.String(d.Get("name").(string))
	}

	if d.HasChange("url") {
		options.Url = scalr.String(d.Get("url").(string))
	}

	if d.HasChange("enabled") {
		options.Enabled = scalr.Bool(d.Get("enabled").(bool))
	}

	if d.HasChange("secret_key") {
		options.SecretKey = scalr.String(d.Get("secret_key").(string))
	}

	if d.HasChange("timeout") {
		options.Timeout = scalr.Int(d.Get("timeout").(int))
	}

	if d.HasChange("max_attempts") {
		options.MaxAttempts = scalr.Int(d.Get("max_attempts").(int))
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
			options.IsShared = scalr.Bool(true)
			options.Environments = make([]*scalr.Environment, 0)
		} else {
			options.IsShared = scalr.Bool(false)
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
		options.IsShared = scalr.Bool(false)
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
	endpointID := d.Get("endpoint_id").(string)

	var err error
	if endpointID != "" {
		err = updateOldWebhook(ctx, d, scalrClient)
	} else {
		err = updateNewWebhook(ctx, d, scalrClient)
	}
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
