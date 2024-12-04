package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
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

func resourceScalrVariableStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	scalrClient := meta.(*scalr.Client)

	humanID := rawState["workspace_id"].(string)
	if !strings.ContainsAny(humanID, "|/") {
		// Due to migration drift, schema-versionV0 can contain workspace_id in format: <WORKSPACE_ID>
		// so we can skip V0->V1 the migration.
		return rawState, nil
	}
	id, err := fetchWorkspaceID(ctx, humanID, scalrClient)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s: %v", humanID, err)
	}

	rawState["workspace_id"] = id
	return rawState, nil
}

func resourceScalrVariableResourceV1() *schema.Resource {
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
						string(scalr.CategoryShell),
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

			"final": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"force": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceScalrVariableStateUpgradeV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	varCategory := rawState["category"].(string)
	if varCategory == string(scalr.CategoryEnv) {
		varCategory = string(scalr.CategoryShell)
	}
	rawState["category"] = varCategory
	return rawState, nil
}

func resourceScalrVariableResourceV2() *schema.Resource {
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
						string(scalr.CategoryShell),
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

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"final": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"force": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceScalrVariableStateUpgradeV2(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	scalrClient := meta.(*scalr.Client)

	varID := rawState["id"].(string)
	//	var, err := scalrClient.variables.ReadByID(varID)
	variable, err := scalrClient.Variables.Read(ctx, varID)
	if err != nil {
		return nil, fmt.Errorf("Error reading variable %s: %v", varID, err)
	}

	rawState["description"] = variable.Description
	return rawState, nil
}
