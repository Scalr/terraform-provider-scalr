package provider

import (
	"context"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrIamUser() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a Scalr user.",
		ReadContext: dataSourceScalrIamUserRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "An identifier of a user.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"email"},
			},
			"status": {
				Description: "A system status of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description:  "An email of a user.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"username": {
				Description: "A username of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"full_name": {
				Description: "A full name of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"identity_providers": {
				Description: "A list of the identity providers the user belongs to.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"teams": {
				Description: "A list of the team identifiers the user belongs to.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrIamUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// required fields
	uID := d.Get("id").(string)
	email := d.Get("email").(string)

	options := scalr.UserListOptions{}

	if uID != "" {
		options.User = ptr(uID)
	}
	if email != "" {
		options.Email = ptr(email)
	}

	log.Printf("[DEBUG] Read configuration of iam user: email '%s', ID '%s'", email, uID)

	ul, err := scalrClient.Users.List(ctx, options)
	if err != nil {
		return diag.Errorf("error retrieving iam user: %v", err)
	}

	if ul.TotalCount == 0 {
		return diag.Errorf("iam user with email '%s' and ID '%s' not found", email, uID)
	}

	u := ul.Items[0]

	// Update the configuration.
	_ = d.Set("email", u.Email)
	_ = d.Set("status", u.Status)
	_ = d.Set("username", u.Username)
	_ = d.Set("full_name", u.FullName)

	var idps []string
	if len(u.IdentityProviders) != 0 {
		for _, idp := range u.IdentityProviders {
			idps = append(idps, idp.ID)
		}
		sort.Strings(idps)
	}
	_ = d.Set("identity_providers", idps)

	var teams []string
	if len(u.Teams) != 0 {
		for _, t := range u.Teams {
			teams = append(teams, t.ID)
		}
		sort.Strings(teams)
	}
	_ = d.Set("teams", teams)

	d.SetId(u.ID)

	return nil
}
