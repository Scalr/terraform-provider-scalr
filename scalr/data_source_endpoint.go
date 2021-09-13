package scalr

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	scalr "github.com/scalr/go-scalr"
)

func dataSourceScalrEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalrEndpointRead,

		Schema: map[string]*schema.Schema{

			"id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"max_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"environment_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceScalrEndpointRead(d *schema.ResourceData, meta interface{}) error {
	scalrClient := meta.(*scalr.Client)

	// Get the ID
	endpointID := d.Get("id").(string)

	log.Printf("[DEBUG] Read endpoint with ID: %s", endpointID)
	endpoint, err := scalrClient.Endpoints.Read(ctx, endpointID)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound{}) {
			return fmt.Errorf("Could not find endpoint %s: %v", endpointID, err)
		}
		return fmt.Errorf("Error retrieving endpoint: %v", err)
	}

	// Update the config.
	d.Set("name", endpoint.Name)
	d.Set("timeout", endpoint.Timeout)
	d.Set("max_attempts", endpoint.MaxAttempts)
	d.Set("secret_key", endpoint.SecretKey)
	d.Set("url", endpoint.Url)
	if endpoint.Environment != nil {
		d.Set("environment_id", endpoint.Environment.ID)
	}
	d.SetId(endpointID)

	return nil
}
