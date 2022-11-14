package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	scalr "github.com/scalr/go-scalr"
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
				Type:     schema.TypeString,
				Optional: true,
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
	var accountID string

	// required fields
	name := d.Get("name").(string)

	options := scalr.TeamListOptions{
		Name: scalr.String(name),
	}
	if accID, ok := d.GetOk("account_id"); ok {
		accountID = accID.(string)
		options.Account = scalr.String(accountID)
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
	d.Set("description", t.Description)
	d.Set("identity_provider_id", t.IdentityProvider.ID)
	if t.Account != nil {
		d.Set("account_id", t.Account.ID)
	}

	var users []string
	if len(t.Users) != 0 {
		for _, u := range t.Users {
			users = append(users, u.ID)
		}
	}
	d.Set("users", users)

	d.SetId(t.ID)

	return nil
}
