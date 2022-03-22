package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrRunTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrRunTriggerCreate,
		Delete: resourceScalrRunTriggerDelete,
		Read:   resourceScalrRunTriggerRead,

		Schema: map[string]*schema.Schema{
			"downstream_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"upstream_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrRunTriggerCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	downstreamID := d.Get("downstream_id").(string)
	upstreamID := d.Get("upstream_id").(string)

	createOptions := scalr.RunTriggerCreateOptions{
		Downstream: &scalr.Downstream{ID: downstreamID},
		Upstream:   &scalr.Upstream{ID: upstreamID},
	}

	log.Printf("[DEBUG] Create run trigger with downstream %s and upstream %s", downstreamID, upstreamID)
	runTrigger, err := scalrClient.RunTriggers.Create(ctx, createOptions)
	if err != nil {
		return fmt.Errorf(
			"Error creating run trigger with downstream %s and upstream %s: %v", downstreamID, upstreamID, err)
	}
	d.SetId(runTrigger.ID)
	return resourceScalrRunTriggerRead(d, meta)

}

func resourceScalrRunTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	log.Printf("[DEBUG] Delete run trigger with ID: %s", id)
	err := scalrClient.RunTriggers.Delete(ctx, id)

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting run trigger %s: %v", id, err)
	}

	return nil
}

func resourceScalrRunTriggerRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Id()

	log.Printf("[DEBUG] Read run trigger %s", id)
	runTrigger, err := scalrClient.RunTriggers.Read(ctx, id)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			log.Printf("[DEBUG] RunTrigger %s no longer exists", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of run trigger %s: %v", id, err)
	}
	d.Set("downstream_id", runTrigger.Downstream.ID)
	d.Set("upstream_id", runTrigger.Upstream.ID)

	return nil
}
