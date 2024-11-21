package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrWorkspaceIDs() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves a map of workspace IDs based on the names provided. Wildcards are accepted.",
		ReadContext: dataSourceScalrWorkspaceIDsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Description: "A list of names to search for. If a name does not exist, it will not throw an error, it will just not exist in the returned output. Use `[\"*\"]` to select all workspaces.",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
			},

			"environment_id": {
				Description: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"ids": {
				Description: "A map of workspace names and their opaque IDs, in the format `env_id/name`.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
		},
	}
}

func dataSourceScalrWorkspaceIDsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	options := scalr.WorkspaceListOptions{
		Filter: &scalr.WorkspaceFilter{Environment: &environmentID},
	}
	for {
		wl, err := scalrClient.Workspaces.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving workspaces: %v", err)
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

	_ = d.Set("ids", ids)
	d.SetId(fmt.Sprintf("%s/%d", environmentID, schema.HashString(id)))

	return nil
}
