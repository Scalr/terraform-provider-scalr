package scalr

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrRoleRead,

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
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},

			"is_system": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceScalrRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// required fields
	name := d.Get("name").(string)
	accountId := d.Get("account_id").(string)

	options := scalr.RoleListOptions{
		Name:    name,
		Account: scalr.String("in:null," + accountId),
	}

	log.Printf("[DEBUG] Read configuration of role: %s/%s", accountId, name)
	roles, err := scalrClient.Roles.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving role: %s/%s", accountId, name)
	}

	// Unlikely situation, but still
	if roles.TotalCount > 1 {
		return diag.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if roles.TotalCount == 0 {
		return diag.Errorf("Could not find role %s/%s", accountId, name)
	}

	role := roles.Items[0]

	// Update the config.
	_ = d.Set("id", role.ID)
	_ = d.Set("is_system", role.IsSystem)
	_ = d.Set("description", role.Description)
	d.SetId(role.ID)

	if len(role.Permissions) != 0 {
		permissionNames := make([]string, 0)

		for _, permission := range role.Permissions {
			permissionNames = append(permissionNames, permission.ID)
		}
		_ = d.Set("permissions", permissionNames)
	}

	if role.Account == nil {
		_ = d.Set("account_id", nil)
	}

	return nil
}
