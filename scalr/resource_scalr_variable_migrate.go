package scalr

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrVariableResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},

			"value": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},

			"category": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(scalr.CategoryEnv),
						string(scalr.CategoryTerraform),
					},
					false,
				),
			},

			"hcl": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"sensitive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceScalrVariableStateUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	scalrClient := meta.(*scalr.Client)

	humanID := rawState["workspace_id"].(string)
	if !strings.ContainsAny(humanID, "|/") {
		// In some obscure cases schema-versionV0 can contain workspace_id in format: <WORKSPACE_ID>
		// so we skip the migration for these cases.
		return rawState, nil
	}
	id, err := fetchWorkspaceID(humanID, scalrClient)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s: %v", humanID, err)
	}

	rawState["workspace_id"] = id
	return rawState, nil
}
