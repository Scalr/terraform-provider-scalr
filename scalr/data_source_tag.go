package scalr

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scalr/go-scalr"
	"log"
)

func dataSourceScalrTag() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrTagRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceScalrTagRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the name and account_id.
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	log.Printf("[DEBUG] Read tag: %s", name)
	tag, err := scalrClient.Tags.Read(ctx, accountID, name)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return fmt.Errorf("Could not find tag %s: %v", name, err)
		}
		return fmt.Errorf("Error retrieving tag: %v", err)
	}

	d.SetId(tag.ID)

	return nil
}
