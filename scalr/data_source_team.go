package scalr

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceTFETeam() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTFETeamRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTFETeamRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create an options struct.
	options := scalr.TeamListOptions{}

	for {
		l, err := scalrClient.Teams.List(ctx, organization, options)
		if err != nil {
			return fmt.Errorf("Error retrieving teams: %v", err)
		}

		for _, tm := range l.Items {
			if tm.Name == name {
				d.SetId(tm.ID)
				return nil
			}
		}

		// Exit the loop when we've seen all pages.
		if l.CurrentPage >= l.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = l.NextPage
	}

	return fmt.Errorf("Could not find team %s/%s", organization, name)
}
