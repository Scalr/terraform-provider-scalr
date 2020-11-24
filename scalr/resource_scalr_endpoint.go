package scalr

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func resourceScalrEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalrEndpointCreate,
		Read:   resourceScalrEndpointRead,
		Update: resourceScalrEndpointUpdate,
		Delete: resourceScalrEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceScalrEndpointResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrVariableStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"max_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceScalrEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get attributes.
	name := d.Get("name").(string)

	// Get scope
	environmentID := d.Get("environment_id").(string)
	// we don't create endpoints on workspace scope for now
	_, environment, account, err := getResourceScope(scalrClient, "", environmentID)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := scalr.EndpointCreateOptions{
		Name:        scalr.String(name),
		SecretKey:   scalr.String(d.Get("secret_key").(string)),
		Url:         scalr.String(d.Get("url").(string)),
		Environment: environment,
		Account:     account,
	}

	if maxAttempts, ok := d.GetOk("max_attempts"); ok {
		options.MaxAttempts = scalr.Int(maxAttempts.(int))
	}

	if timeout, ok := d.GetOk("timeout"); ok {
		options.Timeout = scalr.Int(timeout.(int))
	}

	log.Printf("[DEBUG] Create endpoint: %s", name)
	endpoint, err := scalrClient.Endpoints.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating endpoint %s: %v", name, err)
	}

	d.SetId(endpoint.ID)

	return resourceScalrEndpointRead(d, meta)
}

func resourceScalrEndpointRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)
	endpointID := d.Id()

	log.Printf("[DEBUG] Read endpoint with ID: %s", endpointID)
	endpoint, err := scalrClient.Endpoints.Read(ctx, endpointID)
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return fmt.Errorf("Could not find endpoint %s: %v", endpointID, err)
		}
		return fmt.Errorf("Error retrieving endpoint: %v", err)
	}

	// Update the config.
	d.Set("name", endpoint.Name)
	d.Set("timeout", endpoint.Timeout)
	d.Set("max_attempts", endpoint.MaxAttempts)
	d.Set("secret_key", endpoint.SecretKey)
	if endpoint.Environment != nil {
		d.Set("environment_id", endpoint.Environment.ID)
	}
	d.SetId(endpointID)

	return nil
}

func resourceScalrEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	var err error
	// Create a new options struct.
	options := scalr.EndpointUpdateOptions{
		Name:      scalr.String(d.Get("name").(string)),
		Url:       scalr.String(d.Get("url").(string)),
		SecretKey: scalr.String(d.Get("secret_key").(string)),
	}

	if maxAttempts, ok := d.GetOk("max_attempts"); ok {
		options.MaxAttempts = scalr.Int(maxAttempts.(int))
	}

	if timeout, ok := d.GetOk("timeout"); ok {
		options.Timeout = scalr.Int(timeout.(int))
	}

	log.Printf("[DEBUG] Update endpoint: %s", d.Id())
	_, err = scalrClient.Endpoints.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating endpoint %s: %v", d.Id(), err)
	}

	return resourceScalrEndpointRead(d, meta)
}

func resourceScalrEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete endpoint: %s", d.Id())
	err := scalrClient.Endpoints.Delete(ctx, d.Id())
	if err != nil {
		if err == scalr.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting endpoint%s: %v", d.Id(), err)
	}

	return nil
}
