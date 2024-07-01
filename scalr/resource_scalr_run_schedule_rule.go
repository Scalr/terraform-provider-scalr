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

func resourceScalrRunScheduleRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the state of run schedule rules in Scalr.",
		CreateContext: resourceScalrRunScheduleRuleCreate,
		ReadContext:   resourceScalrRunScheduleRuleRead,
		UpdateContext: resourceScalrRunScheduleRuleUpdate,
		DeleteContext: resourceScalrRunScheduleRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"schedule": {
				Description: "Cron expression for scheduled runs. Time should be in UTC.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"schedule_mode": {
				Description: "Mode of the scheduled run (\"apply\", \"destroy\", \"refresh\").",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					string(scalr.ScheduleModeApply),
					string(scalr.ScheduleModeDestroy),
					string(scalr.ScheduleModeRefresh),
				}, false),
			},
			"workspace_id": {
				Description: "The identifier of the Scalr workspace, in the format `ws-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceScalrRunScheduleRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Read run schedule rule: %s", id)
	rule, err := scalrClient.RunScheduleRules.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] Run schedule rule %s not found", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading run schedule rule %s: %v", id, err)
	}

	// Update config.
	_ = d.Set("schedule", rule.Schedule)
	_ = d.Set("schedule_mode", rule.ScheduleMode)
	_ = d.Set("workspace_id", rule.Workspace.ID)
	return nil
}

func resourceScalrRunScheduleRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Create a new options struct.
	options := scalr.RunScheduleRuleCreateOptions{
		Schedule:     d.Get("schedule").(string),
		ScheduleMode: scalr.ScheduleMode(d.Get("schedule_mode").(string)),
	}

	options.Workspace = &scalr.Workspace{ID: d.Get("workspace_id").(string)}

	log.Printf("[DEBUG] Create run schedule rule %s %s for workspace %s", options.ScheduleMode, options.Schedule, options.Workspace.ID)
	rule, err := scalrClient.RunScheduleRules.Create(ctx, options)
	if err != nil {
		return diag.Errorf(
			"Error creating run schedule rule %s %s for workspace %s: %v", options.ScheduleMode, options.Schedule, options.Workspace.ID, err)
	}
	d.SetId(rule.ID)

	return resourceScalrRunScheduleRuleRead(ctx, d, meta)
}

func resourceScalrRunScheduleRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()
	if d.HasChange("schedule") || d.HasChange("schedule_mode") {
		opts := scalr.RunScheduleRuleUpdateOptions{}

		if v, ok := d.GetOk("schedule"); ok {
			opts.Schedule = scalr.String(v.(string))
		}

		if v, ok := d.GetOk("schedule_mode"); ok {
			mode := scalr.ScheduleMode(v.(string))
			opts.ScheduleMode = &mode
		}

		log.Printf("[DEBUG] Update run schedule rule %s", id)
		_, err := scalrClient.RunScheduleRules.Update(ctx, id, opts)
		if err != nil {
			return diag.Errorf("Error updating run schedule rule %s: %v", id, err)
		}
	}

	return resourceScalrRunScheduleRuleRead(ctx, d, meta)
}

func resourceScalrRunScheduleRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	id := d.Id()

	log.Printf("[DEBUG] Delete run schedule rule %s", id)
	err := scalrClient.RunScheduleRules.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting run schedule rule %s: %v", id, err)
	}

	return nil
}
