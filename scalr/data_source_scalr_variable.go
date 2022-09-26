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
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// computed attributes
			"hcl": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"sensitive": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"final": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		}}
}

func dataSourceScalrVariableRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	filters := scalr.VariableFilter{}
	options := scalr.VariableListOptions{Filter: &filters}

	if keyI, ok := d.GetOk("key"); ok {
		filters.Key = scalr.String(keyI.(string))
	}

	if categoryI, ok := d.GetOk("category"); ok {
		filters.Category = scalr.String(categoryI.(string))
	}

	if accountI, ok := d.GetOk("account_id"); ok {
		filters.Account = scalr.String(accountI.(string))
	}

	if envIdI, ok := d.GetOk("environment_id"); ok {
		filters.Environment = scalr.String(envIdI.(string))
	}
	if workspaceIDI, ok := d.GetOk("workspace_id"); ok {
		filters.Workspace = scalr.String(workspaceIDI.(string))
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

	d.SetId(variable.ID)

	if variable.Environment != nil {
		d.Set("environment_id", variable.Environment.ID)
	}
	if variable.Workspace != nil {
		d.Set("workspace_id", variable.Workspace.ID)
	}

	d.Set("hcl", variable.HCL)
	d.Set("sensitive", variable.Sensitive)
	d.Set("final", variable.Final)
	d.Set("value", variable.Value)
	d.Set("description", variable.Description)

	return nil
}