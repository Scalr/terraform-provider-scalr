package scalr

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrIamTeams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrIamTeamsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
            "identity_provider_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrIamTeamsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	accountID := d.Get("account_id").(string)
	name := d.Get("name").(string)

	options := scalr.TeamListOptions{
		Name:    &name,
		Account: scalr.String("in:null," + accountID),
	}

	var ids []string

	for {
	    tl, err := scalrClient.Teams.List(ctx, options)
        if err != nil {
            return diag.Errorf("Error retrieving iam team: %v", err)
        }
        for _, team := range tl.Items {
			ids = append(ids, team.ID)
		}

		// Exit the loop when we've seen all pages.
		if tl.CurrentPage >= tl.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = tl.NextPage
	}

	_ = d.Set("ids", ids)
	d.SetId(fmt.Sprintf("%d", schema.HashString(accountID+name)))

	return nil
}
