package scalr

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrWorkspaceIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrWorkspaceIDsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ids": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceScalrWorkspaceIDsRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the environment_id.
	environmentID := d.Get("environment_id").(string)

	// Create a map with all the names we are looking for.
	var id string
	names := make(map[string]bool)
	for _, name := range d.Get("names").([]interface{}) {
		id += name.(string)
		names[name.(string)] = true
	}

	// Create a map to store workspace IDs
	ids := make(map[string]string, len(names))

	options := scalr.WorkspaceListOptions{Environment: &environmentID}
	for {
		wl, err := scalrClient.Workspaces.List(ctx, options)
		if err != nil {
			return fmt.Errorf("Error retrieving workspaces: %v", err)
		}

		for _, w := range wl.Items {
			if names["*"] || names[w.Name] {
				ids[w.Name] = w.ID
			}
		}

		// Exit the loop when we've seen all pages.
		if wl.CurrentPage >= wl.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = wl.NextPage
	}

	d.Set("ids", ids)
	d.SetId(fmt.Sprintf("%s/%d", environmentID, schema.HashString(id)))

	return nil
}
