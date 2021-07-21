package scalr

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrIamRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrIamRoleRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceScalrIamRoleRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	id := d.Get("id").(string)

	log.Printf("[DEBUG] Read configuration of role: %s", id)
	role, err := scalrClient.Roles.Read(ctx, id)
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return fmt.Errorf("Could not find role with ID %s", id)
		}
		return fmt.Errorf("Error retrieving role: %v", err)
	}

	// Update the config.
	d.Set("name", role.Name)
	d.Set("description", role.Description)
	d.Set("account_id", role.Account.ID)
	d.SetId(role.ID)

	if len(role.Permissions) != 0 {
		permissionNames := make([]string, 0)

		for _, permission := range role.Permissions {
			permissionNames = append(permissionNames, permission.ID)
		}
		d.Set("permissions", permissionNames)
	}
	d.Set("is_system", role.IsSystem)

	return nil
}
