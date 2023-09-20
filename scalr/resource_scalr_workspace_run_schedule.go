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
		Description:   "Allows workspace admins to automate the configuration of recurring runs for a workspace.",
		CreateContext: resourceScalrWorkspaceRunScheduleCreate,
		ReadContext:   resourceScalrWorkspaceRunScheduleRead,
		UpdateContext: resourceScalrWorkspaceRunScheduleUpdate,
		DeleteContext: resourceScalrWorkspaceRunScheduleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceScalrWorkspaceRunScheduleImport,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this resource. Equals to the ID of the workspace.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"workspace_id": {
				Description: "ID of the workspace, in the format `ws-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"apply_schedule": {
				Description: "Cron expression for when apply run should be created.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"destroy_schedule": {
				Description: "Cron expression for when destroy run should be created.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
		},
	}
}

func resourceScalrWorkspaceRunScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	workspaceId := d.Get("workspace_id").(string)

	// Create a new options struct.
	options := scalr.WorkspaceRunScheduleOptions{}

	if applySchedule, ok := d.GetOk("apply_schedule"); ok {
		options.ApplySchedule = scalr.String(applySchedule.(string))
	}
	if destroySchedule, ok := d.GetOk("destroy_schedule"); ok {
		options.DestroySchedule = scalr.String(destroySchedule.(string))
	}

	applySchedule := ""
	if options.ApplySchedule != nil {
		applySchedule = *options.ApplySchedule
	}

	destroySchedule := ""
	if options.DestroySchedule != nil {
		destroySchedule = *options.DestroySchedule
	}

	log.Printf(
		"[DEBUG] Setting run schedules for workspace ID: %s, apply: %s, destroy: %s",
		workspaceId,
		applySchedule,
		destroySchedule,
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

		if applySchedule, ok := d.GetOk("apply_schedule"); ok {
			options.ApplySchedule = scalr.String(applySchedule.(string))
		}
		if destroySchedule, ok := d.GetOk("destroy_schedule"); ok {
			options.DestroySchedule = scalr.String(destroySchedule.(string))
		}

		applySchedule := ""
		if options.ApplySchedule != nil {
			applySchedule = *options.ApplySchedule
		}

		destroySchedule := ""
		if options.DestroySchedule != nil {
			destroySchedule = *options.DestroySchedule
		}
		log.Printf(
			"[DEBUG] Setting run schedules for workspace ID: %s, apply: %s, destroy: %s",
			workspaceId,
			applySchedule,
			destroySchedule,
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
		ApplySchedule:   nil,
		DestroySchedule: nil,
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
