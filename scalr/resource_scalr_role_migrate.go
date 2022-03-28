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
	for _, perm := range rawState["permissions"].([]interface{}) {
		permissionsSet[perm.(string)] = true
	}

	if permissionsSet["accounts:set-quotas"] {
		return rawState, nil
	}

	if permissionsSet["global-scope:read"] && permissionsSet["accounts:update"] {
		permissionsSet["accounts:set-quotas"] = true
	}

	permissions := make([]string, 0)
	for perm := range permissionsSet {
		permissions = append(permissions, perm)
	}

	// For some reason array should be reverse sorted to produce correct diff ¯\_(ツ)_/¯
	sort.Sort(sort.Reverse(sort.StringSlice(permissions)))

	perms := make([]interface{}, len(permissions))
	for i, perm := range permissions {
		perms[i] = perm
	}

	rawState["permissions"] = perms

	return rawState, nil
}
