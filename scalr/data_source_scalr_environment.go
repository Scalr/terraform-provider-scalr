package scalr

import (
	"context"
	"errors"
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
				AtLeastOneOf: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
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
			"cloud_credentials": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
	environmentName := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	var environment *scalr.Environment
	var err error

	if envID != "" {
		log.Printf("[DEBUG] Read configuration of environment: %s", envID)
		environment, err = scalrClient.Environments.Read(ctx, envID)
	} else {
		log.Printf("[DEBUG] Read configuration of environment: %s", environmentName)
		options := GetEnvironmentByNameOptions{
			Name:    &environmentName,
			Account: &accountID,
			Include: scalr.String("created-by"),
		}
		environment, err = GetEnvironmentByName(ctx, options, scalrClient)
	}

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return diag.Errorf("Environment '%s' not found", envID)
		}
		return diag.Errorf("Error retrieving environment: %v", err)
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
	cloudCredentials := make([]string, 0)
	if environment.CloudCredentials != nil {
		for _, creds := range environment.CloudCredentials {
			cloudCredentials = append(cloudCredentials, creds.ID)
		}
	}
	_ = d.Set("cloud_credentials", cloudCredentials)
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
