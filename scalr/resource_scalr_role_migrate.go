package scalr

import (
	"sort"

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
	for _, perm := range rawState["permissions"].([]string) {
		permissionsSet[perm] = true
	}

	if permissionsSet["accounts:set-quotas"] {
		return rawState, nil
	}
	if permissionsSet["global-scope:*"] || permissionsSet["global-scope:update"] {
		permissionsSet["accounts:set-quotas"] = true
	}

	newPermissions := make([]string, len(permissionsSet))
	counter := 0
	for perm := range permissionsSet {
		newPermissions[counter] = perm
		counter++
	}

	sort.Strings(newPermissions)
	rawState["permissions"] = newPermissions

	return rawState, nil
}
