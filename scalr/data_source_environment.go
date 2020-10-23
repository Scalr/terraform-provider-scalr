package scalr

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrEnvironment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
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

	log.Printf("[DEBUG] Read configuration of environment: %s", envID)

	env, err := scalrClient.Environments.Read(ctx, envID)
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			// If the resource isn't available, the function should set the ID
			// to an empty string so Terraform "destroys" the resource in state.
			d.SetId("")
			return nil
		}
	}
	// Update the config.
	d.Set("name", env.Name)
	d.Set("account_id", env.Account.ID)
	d.Set("cost_estimation_enabled", env.CostEstimationEnabled)
	d.Set("status", env.Status)

	var createdBy []interface{}
	if env.CreatedBy != nil {
		createdBy = append(createdBy, map[string]interface{}{
			"username":  env.CreatedBy.Username,
			"email":     env.CreatedBy.Email,
			"full_name": env.CreatedBy.FullName,
		})
		d.Set("created_by", createdBy)
	}
	cloudCreds := []string{}
	if env.CloudCredentials != nil {
		for _, creds := range env.CloudCredentials {
			cloudCreds = append(cloudCreds, creds.ID)
		}
	}
	d.Set("cloud_credentials", cloudCreds)
	policyGroups := []string{}
	if env.PolicyGroups != nil {
		for _, group := range env.PolicyGroups {
			policyGroups = append(policyGroups, group.ID)
		}
	}
	d.Set("policy_groups", policyGroups)

	d.SetId(envID)

	return nil
}
