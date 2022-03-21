package scalr

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	scalr "github.com/scalr/go-scalr"
	"log"
)

func resourceScalrEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrEnvironmentCreate,
		Read:   resourceScalrEnvironmentRead,
		Delete: resourceScalrEnvironmentDelete,
		Update: resourceScalrEnvironmentUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cost_estimation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
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
				Required: true,
				ForceNew: true,
			},
			"cloud_credentials": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"policy_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func parseCloudCredentialDefinitions(d *schema.ResourceData) ([]*scalr.CloudCredential, error) {
	var cloudCredentials []*scalr.CloudCredential

	cloudCredIds := d.Get("cloud_credentials").([]interface{})
	err := ValidateIDsDefinitions(cloudCredIds)
	if err != nil {
		return nil, fmt.Errorf("Got error during parsing cloud credentials: %s", err.Error())
	}

	for _, cloudCredID := range cloudCredIds {
		cloudCredentials = append(cloudCredentials, &scalr.CloudCredential{ID: cloudCredID.(string)})
	}

	return cloudCredentials, nil
}

func parsePolicyGroupDefinitions(d *schema.ResourceData) ([]*scalr.PolicyGroup, error) {
	var policyGroups []*scalr.PolicyGroup

	policyGroupIds := d.Get("policy_groups").([]interface{})
	err := ValidateIDsDefinitions(policyGroupIds)
	if err != nil {
		return nil, fmt.Errorf("Got error during parsing policy groups: %s", err.Error())
	}

	for _, policyGroupID := range policyGroupIds {
		policyGroups = append(policyGroups, &scalr.PolicyGroup{ID: policyGroupID.(string)})
	}

	return policyGroups, nil
}

func resourceScalrEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)
	cloudCredentials, err := parseCloudCredentialDefinitions(d)
	if err != nil {
		return err
	}
	policyGroups, err := parsePolicyGroupDefinitions(d)
	if err != nil {
		return err
	}

	options := scalr.EnvironmentCreateOptions{
		Name:                  scalr.String(name),
		CostEstimationEnabled: scalr.Bool(d.Get("cost_estimation_enabled").(bool)),
		Account:               &scalr.Account{ID: accountID},
		CloudCredentials:      cloudCredentials,
		PolicyGroups:          policyGroups,
	}
	log.Printf("[DEBUG] Create Environment %s for account: %s", name, accountID)
	environment, err := scalrClient.Environments.Create(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating Environment %s for account %s: %v", name, accountID, err)
	}
	d.SetId(environment.ID)
	return resourceScalrEnvironmentRead(d, meta)
}

func resourceScalrEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	environmentID := d.Id()

	log.Printf("[DEBUG] Read configuration of environment: %s", environmentID)
	environment, err := scalrClient.Environments.Read(ctx, environmentID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			// If the resource isn't available, the function should set the ID
			// to an empty string so Terraform "destroys" the resource in state.
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading environment %s: %v", environmentID, err)
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

	return nil
}

func resourceScalrEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	var err error
	cloudCredentials, err := parseCloudCredentialDefinitions(d)
	if err != nil {
		return err
	}
	policyGroups, err := parsePolicyGroupDefinitions(d)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := scalr.EnvironmentUpdateOptions{
		Name:                  scalr.String(d.Get("name").(string)),
		CostEstimationEnabled: scalr.Bool(d.Get("cost_estimation_enabled").(bool)),
		CloudCredentials:      cloudCredentials,
		PolicyGroups:          policyGroups,
	}
	log.Printf("[DEBUG] Update environment: %s", d.Id())
	_, err = scalrClient.Environments.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating environment %s: %v", d.Id(), err)
	}

	return resourceScalrEnvironmentRead(d, meta)
}

func resourceScalrEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	environmentID := d.Id()

	log.Printf("[DEBUG] Delete environment %s", environmentID)
	err := scalrClient.Environments.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return fmt.Errorf(
			"Error deleting environment %s: %v", environmentID, err)
	}

	return nil
}
