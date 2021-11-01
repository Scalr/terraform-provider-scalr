package scalr

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrRoleRead,

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
				Type:     schema.TypeString,
				Required: true,
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

func dataSourceScalrRoleRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// required fields
	name := d.Get("name").(string)
	accountId := d.Get("account_id").(string)

	options := scalr.RoleListOptions{Name: name, Account: scalr.String(accountId)}
	log.Printf("[DEBUG] Read configuration of role: %s/%s", accountId, name)
	roles, err := scalrClient.Roles.List(ctx, options)
	if err != nil {
		return fmt.Errorf("Error retrieving role: %s/%s", accountId, name)
	}

	// Unlikely situation, but still
	if roles.TotalCount > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if roles.TotalCount == 0 {
		return fmt.Errorf("Could not find role %s/%s", accountId, name)
	}

	role := roles.Items[0]

	// Update the config.
	d.Set("id", role.ID)
	d.Set("is_system", role.IsSystem)
	d.Set("description", role.Description)
	d.SetId(role.ID)

	if len(role.Permissions) != 0 {
		permissionNames := make([]string, 0)

		for _, permission := range role.Permissions {
			permissionNames = append(permissionNames, permission.ID)
		}
		d.Set("permissions", permissionNames)
	}

	return nil
}
