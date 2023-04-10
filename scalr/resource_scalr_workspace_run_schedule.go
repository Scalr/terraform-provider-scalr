package scalr

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
	"log"
)

func resourceScalrWorkspaceRunSchedule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalrWorkspaceRunScheduleCreate,
		ReadContext:   resourceScalrWorkspaceRunScheduleRead,
		UpdateContext: resourceScalrWorkspaceRunScheduleUpdate,
		DeleteContext: resourceScalrWorkspaceRunScheduleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceScalrWorkspaceRunScheduleImport,
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

func resourceScalrWorkspaceRunScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("Error setting run schedule for workspace %s: %v", workspaceId, err)
	}

	d.SetId(workspace.ID)

	return resourceScalrWorkspaceRunScheduleRead(ctx, d, meta)
}

func resourceScalrWorkspaceRunScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	workspaceId := d.Id()

	log.Printf("[DEBUG] Read Workspace with ID: %s", workspaceId)
	workspace, err := scalrClient.Workspaces.ReadByID(ctx, workspaceId)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error retrieving workspace: %v", err)
	}

	// Update the config.
	_ = d.Set("apply_schedule", workspace.ApplySchedule)
	_ = d.Set("destroy_schedule", workspace.DestroySchedule)

	d.SetId(workspace.ID)

	return nil
}

func resourceScalrWorkspaceRunScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
			return diag.Errorf("Error setting run schedule for workspace %s: %v", workspaceId, err)
		}
	}

	return resourceScalrWorkspaceRunScheduleRead(ctx, d, meta)
}

func resourceScalrWorkspaceRunScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("Error deleting workspace run schedules %s: %v", d.Id(), err)
	}

	return nil
}

func resourceScalrWorkspaceRunScheduleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := resourceScalrWorkspaceRunScheduleRead(ctx, d, meta)

	if err != nil {
		return nil, fmt.Errorf("error retrieving workspace run schedule %s: %v", d.Id(), err)
	}

	return []*schema.ResourceData{d}, nil
}
