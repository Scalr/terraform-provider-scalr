package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrVariable() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrVariableRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"key"},
			},
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

func dataSourceScalrVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	filters := scalr.VariableFilter{}
	options := scalr.VariableListOptions{Filter: &filters}

	variableID := d.Get("id").(string)
	key := d.Get("key").(string)

	filters.Account = scalr.String(d.Get("account_id").(string))

	if variableID != "" {
		filters.ID = scalr.String(variableID)
	}
	if key != "" {
		filters.Key = scalr.String(key)
	}
	if categoryI, ok := d.GetOk("category"); ok {
		filters.Category = scalr.String(categoryI.(string))
	}
	if envIdI, ok := d.GetOk("environment_id"); ok {
		filters.Environment = scalr.String(envIdI.(string))
	}
	if workspaceIDI, ok := d.GetOk("workspace_id"); ok {
		filters.Workspace = scalr.String(workspaceIDI.(string))
	}

	variables, err := scalrClient.Variables.List(ctx, options)
	if err != nil {
		return diag.Errorf("Error retrieving Scalr variable: %s.", err)
	}

	if variables.TotalCount > 1 {
		return diag.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	if variables.TotalCount == 0 {
		return diag.Errorf("Could not find a Scalr variable matching you query.")
	}

	variable := variables.Items[0]

	d.SetId(variable.ID)

	if variable.Environment != nil {
		_ = d.Set("environment_id", variable.Environment.ID)
	}
	if variable.Workspace != nil {
		_ = d.Set("workspace_id", variable.Workspace.ID)
	}

	_ = d.Set("key", variable.Key)
	_ = d.Set("category", variable.Category)
	_ = d.Set("hcl", variable.HCL)
	_ = d.Set("sensitive", variable.Sensitive)
	_ = d.Set("final", variable.Final)
	_ = d.Set("value", variable.Value)
	_ = d.Set("description", variable.Description)

	return nil
}
