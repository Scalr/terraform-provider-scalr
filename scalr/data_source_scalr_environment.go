package scalr

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"cost_estimation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"full_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"account_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"policy_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tag_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		}}
}

func dataSourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	envID := d.Get("id").(string)
	envName := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	var environment *scalr.Environment
	var err error

	log.Printf("[DEBUG] Read configuration of environment with ID '%s' and name '%s'", envID, envName)
	if envID != "" {
		environment, err = scalrClient.Environments.Read(ctx, envID)
		if err != nil {
			return diag.Errorf("Error retrieving environment: %v", err)
		}
		if envName != "" && envName != environment.Name {
			return diag.Errorf("Could not find environment with ID '%s' and name '%s'", envID, envName)
		}
	} else {
		options := GetEnvironmentByNameOptions{
			Name:    &envName,
			Account: &accountID,
			Include: scalr.String("created-by"),
		}
		environment, err = GetEnvironmentByName(ctx, options, scalrClient)
		if err != nil {
			return diag.Errorf("Error retrieving environment: %v", err)
		}
		if envID != "" && envID != environment.ID {
			return diag.Errorf("Could not find environment with ID '%s' and name '%s'", envID, envName)
		}
	}

	// Update the configuration.
	_ = d.Set("name", environment.Name)
	_ = d.Set("cost_estimation_enabled", environment.CostEstimationEnabled)
	_ = d.Set("status", environment.Status)

	var createdBy []interface{}
	if environment.CreatedBy != nil {
		createdBy = append(createdBy, map[string]interface{}{
			"username":  environment.CreatedBy.Username,
			"email":     environment.CreatedBy.Email,
			"full_name": environment.CreatedBy.FullName,
		})
	}
	_ = d.Set("created_by", createdBy)
	policyGroups := make([]string, 0)
	if environment.PolicyGroups != nil {
		for _, group := range environment.PolicyGroups {
			policyGroups = append(policyGroups, group.ID)
		}
	}
	_ = d.Set("policy_groups", policyGroups)

	var tags []string
	if len(environment.Tags) != 0 {
		for _, tag := range environment.Tags {
			tags = append(tags, tag.ID)
		}
	}
	_ = d.Set("tag_ids", tags)

	d.SetId(environment.ID)
	return nil
}
