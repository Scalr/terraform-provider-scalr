package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrIamTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrIamTeamRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
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

func dataSourceScalrIamTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// required fields
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.TeamListOptions{
		Name:    &name,
		Account: scalr.String("in:null," + accountID),
	}

	tl, err := scalrClient.Teams.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving iam team: %v", err)
	}

	if tl.TotalCount == 0 {
		return diag.Errorf("Could not find iam team with name %q, account_id: %q", name, accountID)
	}

	if tl.TotalCount > 1 {
		return diag.Errorf(
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
	}

	t := tl.Items[0]

	// Update the configuration.
	_ = d.Set("description", t.Description)
	_ = d.Set("identity_provider_id", t.IdentityProvider.ID)
	if t.Account == nil {
		_ = d.Set("account_id", nil)
	}

	var users []string
	if len(t.Users) != 0 {
		for _, u := range t.Users {
			users = append(users, u.ID)
		}
	}
	_ = d.Set("users", users)

	d.SetId(t.ID)

	return nil
}
