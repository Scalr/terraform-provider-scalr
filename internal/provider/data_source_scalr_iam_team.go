package provider

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrIamTeam() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a Scalr team.",
		ReadContext: dataSourceScalrIamTeamRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "Identifier of the team.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},
			"name": {
				Description:  "Name of the team.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"description": {
				Description: "A verbose description of the team.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"account_id": {
				Description: "The identifier of the Scalr account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"identity_provider_id": {
				Description: "An identifier of an identity provider team is linked to, in the format `idp-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"users": {
				Description: "The list of the user identifiers that belong to the team.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrIamTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// required fields
	teamID := d.Get("id").(string)
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	options := scalr.TeamListOptions{
		Account: ptr("in:null," + accountID),
	}

	if teamID != "" {
		options.Team = ptr(teamID)
	}

	if name != "" {
		options.Name = ptr(name)
	}

	teams, err := scalrClient.Teams.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving iam team: %v", err)
	}

	if teams.TotalCount == 0 {
		return diag.Errorf("Could not find iam team with ID '%s', name '%s', and account_id '%s'", teamID, name, accountID)
	}

	if teams.TotalCount > 1 {
		return diag.Errorf(
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
	}

	team := teams.Items[0]

	// Update the configuration.
	_ = d.Set("name", team.Name)
	_ = d.Set("description", team.Description)
	_ = d.Set("identity_provider_id", team.IdentityProvider.ID)
	if team.Account == nil {
		_ = d.Set("account_id", nil)
	}

	var users []string
	if len(team.Users) != 0 {
		for _, u := range team.Users {
			users = append(users, u.ID)
		}
		sort.Strings(users)
	}
	_ = d.Set("users", users)

	d.SetId(team.ID)

	return nil
}
