package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrVariables() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the list of variables by the given filters.",
		ReadContext: dataSourceScalrVariablesRead,
		Schema: map[string]*schema.Schema{
			"variables": {
				Description: "The list of Scalr variables with all attributes.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the variable.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"category": {
							Description: "Indicates if this is a Terraform or shell variable.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"hcl": {
							Description: "If the variable is configured as a string of HCL code.",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"key": {
							Description: "Key of the variable.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"sensitive": {
							Description: "If the variable is configured as sensitive.",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"final": {
							Description: "If the variable is configured as final. Indicates whether the variable can be overridden on a lower scope down the Scalr organizational model.",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"value": {
							Description: "Variable value if it is not sensitive.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"description": {
							Description: "Variable verbose description.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"workspace_id": {
							Description: "The workspace that owns the variable, specified as an ID, in the format `ws-<RANDOM STRING>`.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"environment_id": {
							Description: "The environment that owns the variable, specified as an ID, in the format `env-<RANDOM STRING>`.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"account_id": {
							Description: "The account that owns the variable, specified as an ID, in the format `acc-<RANDOM STRING>`.",
							Type:        schema.TypeString,
							Optional:    true,
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
					},
				},
				Set: func(value interface{}) int {
					variable := value.(map[string]interface{})

					return schema.HashString(variable["id"].(string))
				},
			},
			"keys": {
				Description: "A list of keys to be used in the query used in a Scalr variable name filter.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"account_id": {
				Description: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"category": {
				Description: "The category of a Scalr variable.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"environment_ids": {
				Description: "A list of identifiers of the Scalr environments, in the format `env-<RANDOM STRING>`. Used to shrink the variable's scope in case the variable name exists in multiple environments.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"workspace_ids": {
				Description: "A list of identifiers of the Scalr workspace, in the format `ws-<RANDOM STRING>`. Used to shrink the variable's scope in case the variable name exists on multiple workspaces.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		}}
}

func dataSourceScalrVariablesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	filters := scalr.VariableFilter{}
	options := scalr.VariableListOptions{Filter: &filters, Include: ptr("updated-by")}

	filters.Account = ptr(d.Get("account_id").(string))

	if keysI, ok := d.GetOk("keys"); ok {
		keys := make([]string, 0)
		for _, keyI := range keysI.(*schema.Set).List() {
			keys = append(keys, keyI.(string))
		}
		if len(keys) > 0 {
			filters.Key = ptr("in:" + strings.Join(keys, ","))
		}
	}
	if categoryI, ok := d.GetOk("category"); ok {
		filters.Category = ptr(categoryI.(string))
	}
	if envIdsI, ok := d.GetOk("environment_ids"); ok {
		envIds := make([]string, 0)
		for _, envIdI := range envIdsI.(*schema.Set).List() {
			envIds = append(envIds, envIdI.(string))
		}
		if len(envIds) > 0 {
			filters.Environment = ptr("in:" + strings.Join(envIds, ","))
		}
	}
	if wsIdsI, ok := d.GetOk("workspace_ids"); ok {
		wsIds := make([]string, 0)
		for _, wsIdI := range wsIdsI.(*schema.Set).List() {
			wsIds = append(wsIds, wsIdI.(string))
		}
		if len(wsIds) > 0 {
			filters.Workspace = ptr("in:" + strings.Join(wsIds, ","))
		}
	}

	variables := make([]map[string]interface{}, 0)
	ids := make([]string, 0)

	for {
		page, err := scalrClient.Variables.List(ctx, options)
		if err != nil {
			return diag.Errorf("Error retrieving Scalr variables: %v", err)
		}

		for _, variable := range page.Items {
			variableI := map[string]interface{}{
				"id":               variable.ID,
				"category":         string(variable.Category),
				"hcl":              variable.HCL,
				"key":              variable.Key,
				"sensitive":        variable.Sensitive,
				"final":            variable.Final,
				"value":            variable.Value,
				"description":      variable.Description,
				"updated_by_email": variable.UpdatedByEmail,
			}
			if variable.UpdatedAt != nil {
				variableI["updated_at"] = variable.UpdatedAt.Format(time.RFC3339)
			}

			var updatedBy []interface{}
			if variable.UpdatedBy != nil {
				updatedBy = append(updatedBy, map[string]interface{}{
					"username":  variable.UpdatedBy.Username,
					"email":     variable.UpdatedBy.Email,
					"full_name": variable.UpdatedBy.FullName,
				})
			}
			variableI["updated_by"] = updatedBy

			if variable.Workspace != nil {
				variableI["workspace_id"] = variable.Workspace.ID
			}
			if variable.Environment != nil {
				variableI["environment_id"] = variable.Environment.ID
			}
			if variable.Account != nil {
				variableI["account_id"] = variable.Account.ID
			}
			variables = append(variables, variableI)
			ids = append(ids, variable.ID)
		}

		if page.CurrentPage >= page.TotalPages {
			break
		}
		options.PageNumber = page.NextPage
	}
	_ = d.Set("variables", variables)
	d.SetId(fmt.Sprintf("%d", schema.HashString(strings.Join(ids, ""))))

	return nil
}
