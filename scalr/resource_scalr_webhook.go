package scalr

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

var (
	eventDefinitions = []string{
		"run:completed",
		"run:errored",
		"run:needs_attention",
	}
)

func init() {
	sort.Strings(eventDefinitions)
}

func resourceScalrWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrWebhookCreate,
		Read:   resourceScalrWebhookRead,
		Update: resourceScalrWebhookUpdate,
		Delete: resourceScalrWebhookDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
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

//remove after https://scalr-labs.atlassian.net/browse/SCALRCORE-16234
func getResourceScope(scalrClient *scalr.Client, workspaceID string, environmentID string) (*scalr.Workspace, *scalr.Environment, *scalr.Account, error) {

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

		if environmentID != "" && environmentID != workspace.Organization.ID {
			return nil, nil, nil, fmt.Errorf("Workspace %s does not belong to an environment %s", workspaceID, environmentID)
		}

		environmentID = workspace.Organization.ID
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

func validateEventDefinitions(eventName string) error {
	if sort.SearchStrings(eventDefinitions, eventName) < len(eventDefinitions) {
		return nil
	}
	eventDefinitionsQuoted := make([]string, len(eventDefinitions))
	for i, eventDefinition := range eventDefinitions {
		eventDefinitionsQuoted[i] = fmt.Sprintf("'%s'", eventDefinition)
	}
	return fmt.Errorf(
		"Invalid value for events '%s'. Allowed values: %s", eventName, strings.Join(eventDefinitionsQuoted, ", "))
}

func resourceScalrWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get attributes.
	name := d.Get("name").(string)
	endpointID := d.Get("endpoint_id").(string)
	workspaceID := d.Get("workspace_id").(string)
	environmentID := d.Get("environment_id").(string)

	workspace, environment, account, err := getResourceScope(scalrClient, workspaceID, environmentID)
	if err != nil {
		return err
	}

	events := d.Get("events").([]interface{})
	var eventDefinitions []*scalr.EventDefinition
	for _, eventID := range events {
		if err := validateEventDefinitions(eventID.(string)); err != nil {
			return err
		}
		eventDefinitions = append(eventDefinitions, &scalr.EventDefinition{ID: eventID.(string)})
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

	log.Printf("[DEBUG] Create webhook: %s", name)
	webhook, err := scalrClient.Webhooks.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating webhook %s: %v", name, err)
	}

	d.SetId(webhook.ID)

	return resourceScalrWebhookRead(d, meta)
}

func resourceScalrWebhookRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the ID
	webhookID := d.Id()

	log.Printf("[DEBUG] Read endpoint with ID: %s", webhookID)
	webhook, err := scalrClient.Webhooks.Read(ctx, webhookID)
	if err != nil {
		if err == scalr.ErrResourceNotFound {
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

	return nil
}

func resourceScalrWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	var err error
	events := d.Get("events").([]string)
	var eventDefinitions []*scalr.EventDefinition
	for _, eventID := range events {
		if err = validateEventDefinitions(eventID); err != nil {
			return err
		}
		eventDefinitions = append(eventDefinitions, &scalr.EventDefinition{ID: eventID})
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

	return resourceScalrWebhookRead(d, meta)
}

func resourceScalrWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete webhook: %s", d.Id())
	err := scalrClient.Webhooks.Delete(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting webhook %s: %v", d.Id(), err)
	}

	return nil
}
