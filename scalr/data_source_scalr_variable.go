package scalr

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrVariable() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a variable.",
		ReadContext: dataSourceScalrVariableRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "ID of a Scalr variable.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"key"},
			},
			"key": {
				Description:  "The name of a Scalr variable.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"category": {
				Description: "The category of a Scalr variable.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"environment_id": {
				Description: "The identifier of the Scalr environment, in the format `env-<RANDOM STRING>`. Used to shrink the scope of the variable in case the variable name exists in multiple environments.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"workspace_id": {
				Description: "The identifier of the Scalr workspace, in the format `ws-<RANDOM STRING>`. Used to shrink the scope of the variable in case the variable name exists on multiple workspaces.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			// computed attributes
			"hcl": {
				Description: "If the variable is configured as a string of HCL code.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"sensitive": {
				Description: "If the variable is configured as sensitive.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"final": {
				Description: "If the variable is configured as final. Indicates whether the variable can be overridden on a lower scope down the Scalr organizational model.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"value": {
				Description: "Variable value.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Variable verbose description, defaults to empty string.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "Date/time the variable was updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_by_email": {
				Description: "Email of the user who updated the variable last time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_by": {
				Description: "Details of the user that updated the variable last time.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Description: "Username of editor.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"email": {
							Description: "Email address of editor.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"full_name": {
							Description: "Full name of editor.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		}}
}

func dataSourceScalrVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	filters := scalr.VariableFilter{}
	options := scalr.VariableListOptions{Filter: &filters, Include: scalr.String("updated-by")}

	variableID := d.Get("id").(string)
	key := d.Get("key").(string)

	filters.Account = scalr.String(d.Get("account_id").(string))

	if variableID != "" {
		filters.Var = scalr.String(variableID)
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
	_ = d.Set("updated_by_email", variable.UpdatedByEmail)

	if variable.UpdatedAt != nil {
		_ = d.Set("updated_at", variable.UpdatedAt.Format(time.RFC3339))
	}

	var updatedBy []interface{}
	if variable.UpdatedBy != nil {
		updatedBy = append(updatedBy, map[string]interface{}{
			"username":  variable.UpdatedBy.Username,
			"email":     variable.UpdatedBy.Email,
			"full_name": variable.UpdatedBy.FullName,
		})
	}
	_ = d.Set("updated_by", updatedBy)

	return nil
}
