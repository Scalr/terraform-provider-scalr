package scalr

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrIamUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrIamUserRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"email"},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"teams": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		options.User = scalr.String(uID)
	}
	if email != "" {
		options.Email = scalr.String(email)
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
	}
	_ = d.Set("identity_providers", idps)

	var teams []string
	if len(u.Teams) != 0 {
		for _, t := range u.Teams {
			teams = append(teams, t.ID)
		}
	}
	_ = d.Set("teams", teams)

	d.SetId(u.ID)

	return nil
}
