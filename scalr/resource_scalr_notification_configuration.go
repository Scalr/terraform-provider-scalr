package scalr

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	scalr "github.com/scalr/go-scalr"
)

func resourceTFENotificationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFENotificationConfigurationCreate,
		Read:   resourceTFENotificationConfigurationRead,
		Update: resourceTFENotificationConfigurationUpdate,
		Delete: resourceTFENotificationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"destination_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.NotificationDestinationTypeGeneric),
						string(scalr.NotificationDestinationTypeSlack),
					},
					false,
				),
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"triggers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						[]string{
							string(scalr.NotificationTriggerCreated),
							string(scalr.NotificationTriggerPlanning),
							string(scalr.NotificationTriggerNeedsAttention),
							string(scalr.NotificationTriggerApplying),
							string(scalr.NotificationTriggerCompleted),
							string(scalr.NotificationTriggerErrored),
						},
						false,
					),
				},
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"workspace_external_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTFENotificationConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get workspace
	workspaceID := d.Get("workspace_external_id").(string)

	// Get attributes
	destinationType := scalr.NotificationDestinationType(d.Get("destination_type").(string))
	enabled := d.Get("enabled").(bool)
	name := d.Get("name").(string)
	token := d.Get("token").(string)
	url := d.Get("url").(string)

	// Throw error if token is set with destinationType of slack
	if token != "" && destinationType == scalr.NotificationDestinationTypeSlack {
		return fmt.Errorf("Token cannot be set with destination_type of %s", destinationType)
	}

	// Create a new options struct
	options := scalr.NotificationConfigurationCreateOptions{
		DestinationType: scalr.NotificationDestination(destinationType),
		Enabled:         scalr.Bool(enabled),
		Name:            scalr.String(name),
		Token:           scalr.String(token),
		URL:             scalr.String(url),
	}

	// Add triggers set to the options struct
	for _, trigger := range d.Get("triggers").(*schema.Set).List() {
		options.Triggers = append(options.Triggers, trigger.(string))
	}

	log.Printf("[DEBUG] Create notification configuration: %s", name)
	notificationConfiguration, err := scalrClient.NotificationConfigurations.Create(ctx, workspaceID, options)
	if err != nil {
		return fmt.Errorf("Error creating notification configuration %s: %v", name, err)
	}

	d.SetId(notificationConfiguration.ID)

	return resourceTFENotificationConfigurationRead(d, meta)
}

func resourceTFENotificationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Read notification configuration: %s", d.Id())
	notificationConfiguration, err := scalrClient.NotificationConfigurations.Read(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			log.Printf("[DEBUG] Notification configuration %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading notification configuration %s: %v", d.Id(), err)
	}

	// Update config
	d.Set("destination_type", notificationConfiguration.DestinationType)
	d.Set("enabled", notificationConfiguration.Enabled)
	d.Set("name", notificationConfiguration.Name)
	// Don't set token here, as it is write only
	// and setting it here would make it blank
	d.Set("triggers", notificationConfiguration.Triggers)
	d.Set("url", notificationConfiguration.URL)

	return nil
}

func resourceTFENotificationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get attributes
	destinationType := scalr.NotificationDestinationType(d.Get("destination_type").(string))
	enabled := d.Get("enabled").(bool)
	name := d.Get("name").(string)
	token := d.Get("token").(string)
	url := d.Get("url").(string)

	// Throw error if token is set with destinationType of slack
	if token != "" && destinationType == scalr.NotificationDestinationTypeSlack {
		return fmt.Errorf("Token cannot be set with destination_type of %s", destinationType)
	}

	// Create a new options struct
	options := scalr.NotificationConfigurationUpdateOptions{
		Enabled: scalr.Bool(enabled),
		Name:    scalr.String(name),
		Token:   scalr.String(token),
		URL:     scalr.String(url),
	}

	// Add triggers set to the options struct
	for _, trigger := range d.Get("triggers").(*schema.Set).List() {
		options.Triggers = append(options.Triggers, trigger.(string))
	}

	log.Printf("[DEBUG] Update notification configuration: %s", d.Id())
	_, err := scalrClient.NotificationConfigurations.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating notification configuration %s: %v", d.Id(), err)
	}

	return resourceTFENotificationConfigurationRead(d, meta)
}

func resourceTFENotificationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete notification configuration: %s", d.Id())
	err := scalrClient.NotificationConfigurations.Delete(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting notification configuration %s: %v", d.Id(), err)
	}

	return nil
}
