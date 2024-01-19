package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
	"log"
)

func resourceScalrSlackIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manage the state of Slack integrations in Scalr. Create, update and destroy." +
			"\n\n-> **Note** Slack workspace should be connected to Scalr account before using this resource.",
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
				Description:      "Name of the Slack integration.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
			"events": {
				Description: "Terraform run events you would like to receive a Slack notifications for. Supported values are `run_approval_required`, `run_success`, `run_errored`.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice(
							[]string{
								scalr.SlackIntegrationEventRunApprovalRequired,
								scalr.SlackIntegrationEventRunSuccess,
								scalr.SlackIntegrationEventRunErrored,
							},
							false,
						),
					),
				},
				Required: true,
				MinItems: 1,
			},
			"channel_id": {
				Description:      "Slack channel ID the event will be sent to.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
			"account_id": {
				Description: "ID of the account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
				ForceNew:    true,
			},
			"environments": {
				Description: "List of environments where events should be triggered.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
				},
				Required: true,
				MinItems: 1,
			},
			"workspaces": {
				Description: "List of workspaces where events should be triggered. Workspaces should be in provided environments. If no workspace is given for a specified environment, events will trigger in all of its workspaces.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
				},
				Optional: true,
			},
		},
	}
}

func parseEvents(d *schema.ResourceData) []string {
	events := d.Get("events").(*schema.Set).List()
	eventValues := make([]string, 0)

	for _, event := range events {
		eventValues = append(eventValues, event.(string))
	}

	return eventValues
}

func parseEnvironments(d *schema.ResourceData) []*scalr.Environment {
	environments := d.Get("environments").(*schema.Set).List()
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
	accountID := d.Get("account_id").(string)

	options := scalr.SlackIntegrationCreateOptions{
		Name:         &name,
		ChannelId:    scalr.String(d.Get("channel_id").(string)),
		Events:       parseEvents(d),
		Account:      &scalr.Account{ID: accountID},
		Environments: parseEnvironments(d),
	}
	workspaces := parseWorkspaces(d)
	if workspaces != nil {
		options.Workspaces = workspaces
	}

	connection, err := scalrClient.SlackIntegrations.GetConnection(ctx, options.Account.ID)
	if err != nil {
		return diag.Errorf("Error creating slack integration %s: %v", name, err)
	}

	if connection.ID == "" {
		return diag.Errorf(
			"Error creating Slack integration: account %s does not have Slack connection configured."+
				" Connect your Slack workspace to Scalr using UI first.",
			accountID,
		)
	}

	options.Connection = connection

	log.Printf("[DEBUG] Create slack integration: %s", name)
	integration, err := scalrClient.SlackIntegrations.Create(ctx, options)
	if err != nil {
		return diag.Errorf("Error creating slack integration %s: %v", name, err)
	}
	d.SetId(integration.ID)

	return resourceScalrSlackIntegrationRead(ctx, d, meta)
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
		events := parseEvents(d)
		options.Events = events
	}

	if d.HasChange("environments") {
		envs := parseEnvironments(d)
		options.Environments = envs
	}

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
