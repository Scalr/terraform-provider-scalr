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

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"http_method": {
				Type:     schema.TypeString,
				Required: true,
			},

			"max_attempts": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"secret_key": {
				Type:     schema.TypeString,
				Required: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceScalrEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get attributes.
	name := d.Get("name").(string)
	httpMethod := d.Get("http_method").(string)

	// Get scope
	workspaceID := d.Get("workspace_id").(string)
	environmentID := d.Get("environment_id").(string)
	workspace, environment, account, err := getResourceScope(scalrClient, workspaceID, environmentID)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := scalr.EndpointCreateOptions{
		Name:        scalr.String(name),
		HTTPMethod:  scalr.String(httpMethod),
		MaxAttempts: scalr.Int(d.Get("max_attempts").(int)),
		SecretKey:   scalr.String(d.Get("secret_key").(string)),
		Timeout:     scalr.Int(d.Get("timeout").(int)),
		Url:         scalr.String(d.Get("url").(string)),
		Workspace:   workspace,
		Environment: environment,
		Account:     account,
	}

	log.Printf("[DEBUG] Create %s endpoint: %s", httpMethod, name)
	endpoint, err := scalrClient.Endpoints.Create(ctx, options)
	if err != nil {
		return fmt.Errorf("Error creating %s endpoint %s: %v", httpMethod, name, err)
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
	d.Set("http_method", endpoint.HTTPMethod)
	d.Set("secret_key", endpoint.SecretKey)
	if endpoint.Workspace != nil {
		d.Set("workspace_id", endpoint.Workspace.ID)
	}
	if endpoint.Environment != nil {
		d.Set("environment_id", endpoint.Environment.ID)
	}
	d.SetId(endpointID)

	return nil
}

func resourceScalrEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get scope
	workspaceID := d.Get("workspace_id").(string)
	environmentID := d.Get("environment_id").(string)
	workspace, environment, account, err := getResourceScope(scalrClient, workspaceID, environmentID)
	if err != nil {
		return err
	}

	// Create a new options struct.
	options := scalr.EndpointUpdateOptions{
		HTTPMethod:  scalr.String(d.Get("http_method").(string)),
		MaxAttempts: scalr.Int(d.Get("max_attempts").(int)),
		Url:         scalr.String(d.Get("url").(string)),
		SecretKey:   scalr.String(d.Get("secret_key").(string)),
		Timeout:     scalr.Int(d.Get("timeout").(int)),
		Workspace:   workspace,
		Environment: environment,
		Account:     account,
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
