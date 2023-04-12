package scalr

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scalr/go-scalr"
)

func resourceScalrEndpoint() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "Resource `scalr_endpoint` is deprecated, please set the endpoint information" +
			" in the `scalr_webhook` resource.",
		CreateContext: resourceScalrEndpointCreate,
		ReadContext:   resourceScalrEndpointRead,
		UpdateContext: resourceScalrEndpointUpdate,
		DeleteContext: resourceScalrEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceScalrEndpointResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceScalrEndpointStateUpgradeV0,
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
				Optional:  true,
				Computed:  true,
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

func resourceScalrEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	// Get attributes.
	name := d.Get("name").(string)

	// Get scope
	environmentID := d.Get("environment_id").(string)
	// we don't create endpoints on workspace scope for now
	_, environment, account, err := getResourceScope(ctx, scalrClient, "", environmentID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create a new options struct.
	options := scalr.EndpointCreateOptions{
		Name:        scalr.String(name),
		Url:         scalr.String(d.Get("url").(string)),
		Environment: environment,
		Account:     account,
	}
	if secretKey, ok := d.GetOk("secret_key"); ok {
		options.SecretKey = scalr.String(secretKey.(string))
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
		return diag.Errorf("Error creating endpoint %s: %v", name, err)
	}

	d.SetId(endpoint.ID)

	return resourceScalrEndpointRead(ctx, d, meta)
}

func resourceScalrEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)
	endpointID := d.Id()

	log.Printf("[DEBUG] Read endpoint with ID: %s", endpointID)
	endpoint, err := scalrClient.Endpoints.Read(ctx, endpointID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error retrieving endpoint: %v", err)
	}

	// Update the config.
	_ = d.Set("name", endpoint.Name)
	_ = d.Set("timeout", endpoint.Timeout)
	_ = d.Set("max_attempts", endpoint.MaxAttempts)
	_ = d.Set("secret_key", endpoint.SecretKey)
	if endpoint.Environment != nil {
		_ = d.Set("environment_id", endpoint.Environment.ID)
	}
	d.SetId(endpointID)

	return nil
}

func resourceScalrEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("Error updating endpoint %s: %v", d.Id(), err)
	}

	return resourceScalrEndpointRead(ctx, d, meta)
}

func resourceScalrEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scalrClient := meta.(*scalr.Client)

	log.Printf("[DEBUG] Delete endpoint: %s", d.Id())
	err := scalrClient.Endpoints.Delete(ctx, d.Id())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			return nil
		}
		return diag.Errorf("Error deleting endpoint%s: %v", d.Id(), err)
	}

	return nil
}
