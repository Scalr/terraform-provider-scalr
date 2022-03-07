package scalr

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceScalrRoleResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"is_system": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"permissions": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 128,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceScalrRoleStateUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	permissionsSet := make(map[string]bool)
	permissions := rawState["permissions"].([]interface{})
	for _, perm := range permissions {
		permissionsSet[perm.(string)] = true
	}

	if permissionsSet["accounts:set-quotas"] {
		return rawState, nil
	}
	if permissionsSet["global-scope:read"] && permissionsSet["accounts:update"] {
		permissions = append(permissions, "accounts:set-quotas")
		rawState["permissions"] = permissions
	}

	return rawState, nil
}
