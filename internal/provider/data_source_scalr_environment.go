package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func dataSourceScalrEnvironment() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves the details of a Scalr environment.",
		ReadContext: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The environment ID, in the format `env-<RANDOM STRING>`.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				AtLeastOneOf: []string{"name"},
			},
			"name": {
				Description:  "Name of the environment.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"cost_estimation_enabled": {
				Description: "Boolean indicates if cost estimation is enabled for the environment.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"status": {
				Description: "The status of an environment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Details of the user that created the environment.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Description: "Username of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"email": {
							Description: "Email address of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"full_name": {
							Description: "Full name of creator.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"account_id": {
				Description: "ID of the environment account, in the format `acc-<RANDOM STRING>`",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DefaultFunc: scalrAccountIDDefaultFunc,
			},
			"policy_groups": {
				Description: "List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tag_ids": {
				Description: "List of tag IDs associated with the environment.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"default_provider_configurations": {
				Description: "List of IDs of provider configurations, used in the environment workspaces by default.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
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
			Include: ptr("created-by"),
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

	defaultProviderConfigurations := make([]string, 0)
	if environment.DefaultProviderConfigurations != nil {
		for _, provider := range environment.DefaultProviderConfigurations {
			defaultProviderConfigurations = append(defaultProviderConfigurations, provider.ID)
		}
	}

	_ = d.Set("default_provider_configurations", defaultProviderConfigurations)

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
