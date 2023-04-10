package scalr

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrVariables() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalrVariablesRead,
		Schema: map[string]*schema.Schema{
			"variables": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"category": {
							Type:     schema.TypeString,
							Required: true,
						},
						"hcl": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sensitive": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"final": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Required: true,
						},
						"workspace_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"environment_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"account_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: func(value interface{}) int {
					variable := value.(map[string]interface{})

					return schema.HashString(variable["id"].(string))
				},
			},
			"keys": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environment_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"workspace_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		}}
}

func dataSourceScalrVariablesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	filters := scalr.VariableFilter{}
	options := scalr.VariableListOptions{Filter: &filters}

	filters.Account = scalr.String(d.Get("account_id").(string))

	if keysI, ok := d.GetOk("keys"); ok {
		keys := make([]string, 0)
		for _, keyI := range keysI.(*schema.Set).List() {
			keys = append(keys, keyI.(string))
		}
		if len(keys) > 0 {
			filters.Key = scalr.String("in:" + strings.Join(keys, ","))
		}
	}
	if categoryI, ok := d.GetOk("category"); ok {
		filters.Category = scalr.String(categoryI.(string))
	}
	if envIdsI, ok := d.GetOk("environment_ids"); ok {
		envIds := make([]string, 0)
		for _, envIdI := range envIdsI.(*schema.Set).List() {
			envIds = append(envIds, envIdI.(string))
		}
		if len(envIds) > 0 {
			filters.Environment = scalr.String("in:" + strings.Join(envIds, ","))
		}
	}
	if wsIdsI, ok := d.GetOk("workspace_ids"); ok {
		wsIds := make([]string, 0)
		for _, wsIdI := range wsIdsI.(*schema.Set).List() {
			wsIds = append(wsIds, wsIdI.(string))
		}
		if len(wsIds) > 0 {
			filters.Workspace = scalr.String("in:" + strings.Join(wsIds, ","))
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
				"id":          variable.ID,
				"category":    string(variable.Category),
				"hcl":         variable.HCL,
				"key":         variable.Key,
				"sensitive":   variable.Sensitive,
				"final":       variable.Final,
				"value":       variable.Value,
				"description": variable.Description,
			}
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
