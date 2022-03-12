package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrEnvironment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
		}}
}

func dataSourceEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	envID := d.Get("id").(string)
	environmentName := d.Get("name").(string)

	if envID == "" && environmentName == "" {
		return fmt.Errorf("At least one argument 'id' or 'name' is required, but no definitions was found")
	}

	if envID != "" && environmentName != "" {
		return fmt.Errorf("Attributes 'name' and 'id' can not be set at the same time")
	}

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
			Include: scalr.String("created-by"),
		}
		if accountID != "" {
			options.Account = &accountID
		}
		environment, err = GetEnvironmentByName(options, scalrClient)
	}

	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return fmt.Errorf("Environment %s not found", envID)
		}
		return fmt.Errorf("Error retrieving environment: %v", err)
	}
	// Update the configuration.
	d.Set("name", environment.Name)
	d.Set("account_id", environment.Account.ID)
	d.Set("cost_estimation_enabled", environment.CostEstimationEnabled)
	d.Set("status", environment.Status)

	var createdBy []interface{}
	if environment.CreatedBy != nil {
		createdBy = append(createdBy, map[string]interface{}{
			"username":  environment.CreatedBy.Username,
			"email":     environment.CreatedBy.Email,
			"full_name": environment.CreatedBy.FullName,
		})
	}
	d.Set("created_by", createdBy)
	cloudCredentials := []string{}
	if environment.CloudCredentials != nil {
		for _, creds := range environment.CloudCredentials {
			cloudCredentials = append(cloudCredentials, creds.ID)
		}
	}
	d.Set("cloud_credentials", cloudCredentials)
	policyGroups := []string{}
	if environment.PolicyGroups != nil {
		for _, group := range environment.PolicyGroups {
			policyGroups = append(policyGroups, group.ID)
		}
	}
	d.Set("policy_groups", policyGroups)

	d.SetId(environment.ID)
	return nil
}
