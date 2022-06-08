package scalr

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scalr/go-scalr"
	"log"
)

func resourceScalrWorkspaceRunSchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrWorkspaceRunScheduleCreate,
		Read:   resourceScalrWorkspaceRunScheduleRead,
		Update: resourceScalrWorkspaceRunScheduleUpdate,
		Delete: resourceScalrWorkspaceRunScheduleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceScalrWorkspaceRunScheduleImport,
		},

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"apply_schedule": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"destroy_schedule": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceScalrWorkspaceRunScheduleCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	workspaceId := d.Get("workspace_id").(string)

	// Create a new options struct.
	options := scalr.WorkspaceRunScheduleOptions{}

	options.ApplySchedule = d.Get("apply_schedule").(string)
	options.DestroySchedule = d.Get("destroy_schedule").(string)

	log.Printf(
		"[DEBUG] Setting run schedules for workspace ID: %s, apply: %s, destroy: %s",
		workspaceId,
		options.ApplySchedule,
		options.DestroySchedule,
	)
	workspace, err := scalrClient.Workspaces.SetSchedule(ctx, workspaceId, options)
	if err != nil {
		return fmt.Errorf("Error setting run schedule for workspace %s: %v", workspaceId, err)
	}

	d.SetId(workspace.ID)

	return resourceScalrWorkspaceRunScheduleRead(d, meta)
}

func resourceScalrWorkspaceRunScheduleRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	workspaceId := d.Id()

	log.Printf("[DEBUG] Read Workspace with ID: %s", workspaceId)
	workspace, err := scalrClient.Workspaces.ReadByID(ctx, workspaceId)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving workspace: %v", err)
	}

	// Update the config.
	d.Set("apply_schedule", workspace.ApplySchedule)
	d.Set("destroy_schedule", workspace.DestroySchedule)

	d.SetId(workspace.ID)

	return nil
}

func resourceScalrWorkspaceRunScheduleUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	var err error
	workspaceId := d.Id()

	if d.HasChange("apply_schedule") || d.HasChange("destroy_schedule") {
		// Create a new options struct.
		options := scalr.WorkspaceRunScheduleOptions{}

		options.ApplySchedule = d.Get("apply_schedule").(string)
		options.DestroySchedule = d.Get("destroy_schedule").(string)

		log.Printf(
			"[DEBUG] Setting run schedules for workspace ID: %s, apply: %s, destroy: %s",
			workspaceId,
			options.ApplySchedule,
			options.DestroySchedule,
		)
		_, err = scalrClient.Workspaces.SetSchedule(ctx, workspaceId, options)
		if err != nil {
			return fmt.Errorf("Error setting run schedule for workspace %s: %v", workspaceId, err)
		}
	}

	return resourceScalrWorkspaceRunScheduleRead(d, meta)
}

func resourceScalrWorkspaceRunScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete run schedules for workspace: %s", d.Id())
	_, err := scalrClient.Workspaces.SetSchedule(ctx, d.Id(), scalr.WorkspaceRunScheduleOptions{
		ApplySchedule:   "",
		DestroySchedule: "",
	})
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting workspace run schedules %s: %v", d.Id(), err)
	}

	return nil
}

func resourceScalrWorkspaceRunScheduleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := resourceScalrWorkspaceRunScheduleRead(d, meta)

	if err != nil {
		return nil, fmt.Errorf("error retrieving workspace run schedule %s: %v", d.Id(), err)
	}

	return []*schema.ResourceData{d}, nil
}
