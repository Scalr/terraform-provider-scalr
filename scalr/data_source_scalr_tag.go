package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	options := scalr.TagListOptions{
		Account: scalr.String(accountID),
		Name:    scalr.String(name),
	}

	log.Printf("[DEBUG] Read tag: %s/%s", accountID, name)
	tags, err := scalrClient.Tags.List(ctx, options)
	if err != nil {
		return fmt.Errorf("Error retrieving tag: %v", err)
	}

	// Unlikely
	if tags.TotalCount > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if tags.TotalCount == 0 {
		return fmt.Errorf("Could not find tag %s/%s", accountID, name)
	}

	tag := tags.Items[0]
	d.SetId(tag.ID)

	return nil
}
