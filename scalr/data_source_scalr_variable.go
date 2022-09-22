package scalr

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

package scalr

import (
"fmt"

"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrVariable() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrVariableRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeString,
				Required: true,
			},
			"hcl": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sensitive": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"final": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		}}
}

func dataSourceScalrVariableRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	options := scalr.VariableListOptions{}

	// TODO: communicate with PO if renaming key -> name makes any sense. Seems to be a stupid requirement
	if name, ok := d.GetOk("name"); ok {
		options.Key = scalr.String(name.(string))
	}

	if category, ok := d.GetOk("category"); ok {
		options.Category = scalr.String(category.(string))
	}

	if accountId, ok := d.GetOk("account_id"); ok {
		options.Account = scalr.String(accountId.(string))
	}

	if envId, ok := d.GetOk("environment_id"); ok {
		options.Environment = scalr.String(envId.(string))
	}

	if workspaceID, ok := d.GetOk("workspace_id"); ok {
		options.workspace = scalr.String(workspaceID.(string))
	}

	variables, err := scalrClient.Variables.List(ctx, options)

	if err != nil {
		return fmt.Errorf("Error retrieving Scalr variable: %s.", err)
	}

	if variables.TotalCount > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if variables.TotalCount == 0 {
		return fmt.Errorf("Could not find a Scalr variable matching you query.")
	}

	variable := variables.Items[0]

	// TODO: Update the variable.
	// Update the variable.
	d.SetId(variable.ID)

	return nil
}
