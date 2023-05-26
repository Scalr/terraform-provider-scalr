package scalr

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
	"log"
)

func resourceScalrSlackIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrSlackIntegrationCreate,
		ReadContext:   resourceScalrSlackIntegrationRead,
		UpdateContext: resourceScalrSlackIntegrationUpdate,
		DeleteContext: resourceSlackIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"events": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice(
							[]string{scalr.RunApprovalRequiredEvent, scalr.RunSuccessEvent, scalr.RunErroredEvent},
							false,
						),
					),
				},
				Required: true,
				MinItems: 1,
			},
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"environments": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"workspaces": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func parseEvents(d *schema.ResourceData) ([]string, error) {
	events := make([]string, 0)

	providedEvents := d.Get("events").([]interface{})
	err := ValidateIDsDefinitions(providedEvents)
	if err != nil {
		return nil, fmt.Errorf("Got error during parsing events: %s", err.Error())
	}

	for _, event := range providedEvents {
		e := event.(string)
		events = append(events, e)
	}

	return events, nil
}

func parseEnvironments(d *schema.ResourceData) []*scalr.Environment {
	environmentsI := d.Get("environments")
	environments := environmentsI.(*schema.Set).List()
	environmentValues := make([]*scalr.Environment, 0)
	for _, env := range environments {
		environmentValues = append(environmentValues, &scalr.Environment{ID: env.(string)})
	}

	return environmentValues
}

func parseWorkspaces(d *schema.ResourceData) []*scalr.Workspace {
	workspacesI, ok := d.GetOk("workspaces")
	if !ok {
		return nil
	}
	workspaces := workspacesI.(*schema.Set).List()
	workspaceValues := make([]*scalr.Workspace, 0)
	for _, ws := range workspaces {
		workspaceValues = append(workspaceValues, &scalr.Workspace{ID: ws.(string)})
	}

	return workspaceValues
}

func resourceScalrSlackIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	// Get attributes.
	name := d.Get("name").(string)
	channelId := d.Get("channel_id").(string)
	events, err := parseEvents(d)
	if err != nil {
		return diag.Errorf("Error creating slack integration %s: %v", name, err)
	}

	envs := parseEnvironments(d)

	options := scalr.SlackIntegrationCreateOptions{
		Name:         &name,
		ChannelId:    &channelId,
		Events:       events,
		Account:      &scalr.Account{ID: d.Get("account_id").(string)},
		Environments: envs,
	}
	workspaces := parseWorkspaces(d)
	if workspaces != nil {
		options.Workspaces = workspaces
	}

	connection, err := scalrClient.SlackIntegrations.GetConnection(ctx, options.Account.ID)

	if err != nil || connection.ID == "" {
		return diag.Errorf("Error creating slack integration, account not connected to slack, create slack connection using UI first.")
	}

	options.Connection = connection

	log.Printf("[DEBUG] Create slack integration: %s", name)
	integration, err := scalrClient.SlackIntegrations.Create(ctx, options)
	if err != nil {
		return diag.Errorf("Error creating slack integration %s: %v", name, err)
	}
	d.SetId(integration.ID)

	return nil
}

func resourceScalrSlackIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	integrationID := d.Id()

	log.Printf("[DEBUG] Read slack integration with ID: %s", integrationID)
	slackIntegration, err := scalrClient.SlackIntegrations.Read(ctx, integrationID)
	if err != nil {
		log.Printf("[DEBUG] slack integration %s no longer exists", integrationID)
		d.SetId("")
		return nil
	}
	_ = d.Set("name", slackIntegration.Name)
	_ = d.Set("channel_id", slackIntegration.ChannelId)
	_ = d.Set("events", slackIntegration.Events)
	_ = d.Set("account_id", slackIntegration.Account.ID)

	environmentIDs := make([]string, 0)
	for _, environment := range slackIntegration.Environments {
		environmentIDs = append(environmentIDs, environment.ID)
	}

	_ = d.Set("environments", environmentIDs)

	wsIDs := make([]string, 0)
	for _, ws := range slackIntegration.Workspaces {
		wsIDs = append(wsIDs, ws.ID)
	}
	_ = d.Set("workspaces", wsIDs)

	return nil
}

func resourceScalrSlackIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	options := scalr.SlackIntegrationUpdateOptions{}

	if d.HasChange("name") {
		options.Name = scalr.String(d.Get("name").(string))
	}

	if d.HasChange("channel_id") {
		options.ChannelId = scalr.String(d.Get("channel_id").(string))
	}

	if d.HasChange("events") {
		events, err := parseEvents(d)
		if err != nil {
			return diag.Errorf("Error updating slack integration %s: %v", d.Id(), err)
		}
		options.Events = events
	}

	envs := parseEnvironments(d)
	options.Environments = envs

	workspaces := parseWorkspaces(d)
	if workspaces != nil {
		options.Workspaces = workspaces
	}

	log.Printf("[DEBUG] Update slack integration: %s", d.Id())
	_, err := scalrClient.SlackIntegrations.Update(ctx, d.Id(), options)
	if err != nil {
		return diag.Errorf("Error updating slack integration %s: %v", d.Id(), err)
	}

	return nil
}

func resourceSlackIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete slack integration: %s", d.Id())
	err := scalrClient.SlackIntegrations.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting slack integration %s: %v", d.Id(), err)
	}

	return nil
}
